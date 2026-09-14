package listener

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func StartMQTTClient(ctx context.Context, brokerURL string, svc model.DeviceGatewayService, sessionSvc model.SessionManagerService, deviceCacheSvc model.DeviceCacheService) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID("device-gateway-worker")
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(5 * time.Second)

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		slog.InfoContext(ctx, "Connected to MQTT Broker!")

		// Subscribe to Single Device Topic
		singleTopic := "device/+/data"
		// ⚡ ส่ง cacheSvc เข้าไปใน Handler
		sHandler := singleHandler(ctx, svc, sessionSvc, deviceCacheSvc)
		if token := c.Subscribe(singleTopic, 1, sHandler); token.Wait() && token.Error() != nil {
			slog.ErrorContext(ctx, "Failed to subscribe to MQTT topic", slog.String("error", token.Error().Error()))
		} else {
			slog.InfoContext(ctx, "Subscribed to MQTT topic", slog.String("topic", singleTopic))
		}

		// Subscribe to Group Device Topic
		groupTopic := "device-group/+/data"
		// ⚡ ส่ง cacheSvc เข้าไปใน Handler
		gHandler := groupHandler(ctx, svc, sessionSvc, deviceCacheSvc)
		if token := c.Subscribe(groupTopic, 1, gHandler); token.Wait() && token.Error() != nil {
			slog.ErrorContext(ctx, "Failed to subscribe to group topic", slog.String("error", token.Error().Error()))
		} else {
			slog.InfoContext(ctx, "Subscribed to MQTT topic", slog.String("topic", groupTopic))
		}
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		slog.ErrorContext(ctx, "Failed to connect to MQTT broker", slog.String("error", token.Error().Error()))
		return
	}
	sessionSvc.SetMQTTClient(client)

	// Wait for graceful shutdown
	<-ctx.Done()
	slog.Info("Shutting down MQTT listener...")
	client.Disconnect(250)
}

// --- HANDLERS ---
func singleHandler(ctx context.Context, svc model.DeviceGatewayService, sessionSvc model.SessionManagerService, deviceCacheSvc model.DeviceCacheService) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		topicParts := strings.Split(msg.Topic(), "/")
		if len(topicParts) != 3 {
			return
		}
		deviceName := topicParts[1]

		// 1. Resolve Name to ID using Central Cache Service
		deviceId, protocol, err := deviceCacheSvc.GetDeviceInfoByName(ctx, deviceName)
		if err != nil || deviceId <= 0 || protocol != "MQTT" {
			slog.Warn("Unknown device name received", slog.String("name", deviceName))
			return
		}

		payload := bytes.TrimSpace(msg.Payload())
		if len(payload) == 0 {
			return
		}

		var incomingData []model.DeviceDataPayloadReq
		switch payload[0] {
		case '[':
			if err := json.Unmarshal(payload, &incomingData); err != nil {
				slog.ErrorContext(ctx, "Error", slog.String("track", err.Error()))
				return
			}
		case '{':
			var single model.DeviceDataPayloadReq
			if err := json.Unmarshal(payload, &single); err != nil {
				slog.ErrorContext(ctx, "Error", slog.String("track", err.Error()))
				return
			}
			incomingData = append(incomingData, single)
		default:
			return
		}

		// Process payload for this specific device
		for _, d := range incomingData {
			deviceData := model.DeviceData{
				DeviceId:   deviceId,
				DeviceName: deviceName,
				ValueData:  int(math.Round(d.ValueData * model.DeviceScale)),
			}
			sessionSvc.MarkDeviceActive(deviceData.DeviceId)
			svc.Add(deviceData)
		}
	}
}

func groupHandler(ctx context.Context, svc model.DeviceGatewayService, sessionSvc model.SessionManagerService, deviceCacheSvc model.DeviceCacheService) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		topicParts := strings.Split(msg.Topic(), "/")
		if len(topicParts) != 3 {
			return
		}
		groupName := topicParts[1]

		deviceIds, groupProtocol, err := deviceCacheSvc.GetGroupInfoByName(ctx, groupName)
		if err != nil || len(deviceIds) == 0 {
			slog.Warn("Unknown or empty group name received", slog.String("groupName", groupName))
			return
		}

		if groupProtocol != "MQTT" {
			slog.Warn("Group protocol mismatch", slog.String("groupName", groupName), slog.String("expected", "MQTT"), slog.String("got", groupProtocol))
			return
		}

		payload := bytes.TrimSpace(msg.Payload())
		if len(payload) == 0 {
			return
		}

		var incomingData []model.DeviceDataPayloadReq
		switch payload[0] {
		case '[':
			if err := json.Unmarshal(payload, &incomingData); err != nil {
				slog.ErrorContext(ctx, "Error", slog.String("track", err.Error()))
				return
			}
		case '{':
			var single model.DeviceDataPayloadReq
			if err := json.Unmarshal(payload, &single); err != nil {
				slog.ErrorContext(ctx, "Error", slog.String("track", err.Error()))
				return
			}
			incomingData = append(incomingData, single)
		default:
			return
		}

		// ⚡ 2. Map valid devices by ID, not by Name
		validDevices := make(map[int]bool)
		for _, id := range deviceIds {
			validDevices[id] = true
		}

		// Loop through the incoming JSON payload array
		for _, d := range incomingData {
			if d.DeviceName == "" {
				slog.Warn("Device name in payload is missing", slog.String("group", groupName))
				continue
			}

			// ⚡ 3. Resolve Name to ID FIRST using the device cache (which you mentioned is already updating correctly)
			deviceId, _, err := deviceCacheSvc.GetDeviceInfoByName(ctx, d.DeviceName)
			if err != nil || deviceId <= 0 {
				slog.Warn("Unknown device in group payload", slog.String("device", d.DeviceName))
				continue
			}

			// ⚡ 4. Validate if the resolved ID belongs to this group
			if !validDevices[deviceId] {
				slog.Warn("Device does not belong to group",
					slog.String("device", d.DeviceName),
					slog.Int("deviceId", deviceId),
					slog.String("group", groupName),
				)
				continue
			}

			// 5. Save the data
			deviceData := model.DeviceData{
				DeviceId:   deviceId,
				DeviceName: d.DeviceName,
				ValueData:  int(math.Round(d.ValueData * model.DeviceScale)),
			}
			sessionSvc.MarkDeviceActive(deviceData.DeviceId)
			svc.Add(deviceData)
		}
	}
}
