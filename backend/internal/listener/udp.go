package listener

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net"

	"github.com/FTATM/software-dashboard-tool/internal/model"
)

func StartUDPServer(ctx context.Context, port string, svc model.DeviceGatewayService, sessionSvc model.SessionManagerService, deviceCacheSvc model.DeviceCacheService) {
	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to resolve UDP address", slog.String("error", err.Error()))
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to start UDP listener", slog.String("error", err.Error()))
		return
	}
	slog.InfoContext(ctx, "UDP Listener started", slog.String("port", port))

	// Register UDP Connection into SessionManager for routing commands
	sessionSvc.SetUDPServer(conn)

	go func() {
		<-ctx.Done()
		slog.Info("Shutting down UDP listener...")
		conn.Close()
	}()

	// ⚡ 1. Create a map to track known devices and their current UDP address
	knownDevices := make(map[int]string)

	buffer := make([]byte, 2048) // Buffer size for incoming packets

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.WarnContext(ctx, "UDP read error", slog.String("error", err.Error()))
			continue
		}

		packetData := bytes.TrimSpace(buffer[:n])
		if len(packetData) == 0 {
			continue
		}

		var incomingData []model.DeviceDataPayloadReq

		// Check payload format (Array or Object)
		switch packetData[0] {
		case '[':
			if err := json.Unmarshal(packetData, &incomingData); err != nil {
				slog.Debug("Invalid JSON array received over UDP", slog.String("ip", remoteAddr.String()))
				continue
			}
		case '{':
			var single model.DeviceDataPayloadReq
			if err := json.Unmarshal(packetData, &single); err != nil {
				slog.Debug("Invalid JSON object received over UDP", slog.String("ip", remoteAddr.String()))
				continue
			}
			incomingData = append(incomingData, single)
		default:
			continue
		}

		currentAddrString := remoteAddr.String()

		// Process each item in the payload
		for _, req := range incomingData {
			if req.DeviceName == "" {
				continue
			}

			// Find Device ID from Cache Service
			deviceId, protocol, err := deviceCacheSvc.GetDeviceInfoByName(ctx, req.DeviceName)
			if err != nil || deviceId <= 0 {
				slog.Warn("Unknown device in UDP payload", slog.String("device", req.DeviceName))
				continue
			}

			if protocol != "UDP" {
				slog.Warn("Device protocol mismatch", slog.String("DeviceName", req.DeviceName), slog.String("expected", "UDP"), slog.String("got", protocol))
				continue
			}

			// ⚡ 2. Check if this is a new connection or if the device's port changed
			if knownDevices[deviceId] != currentAddrString {
				knownDevices[deviceId] = currentAddrString

				// Register IP/Port to send Commands back
				sessionSvc.RegisterUDP(deviceId, remoteAddr)

				// ⚡ 3. Send ACK back to the UDP client
				ackMsg := fmt.Appendf(nil, "UDP %s connected!\n", req.DeviceName)
				if _, err := conn.WriteToUDP(ackMsg, remoteAddr); err != nil {
					slog.WarnContext(ctx, "Failed to send ACK to UDP device",
						slog.String("device", req.DeviceName),
						slog.String("error", err.Error()),
					)
				}
			}

			// Mark device as online
			sessionSvc.MarkDeviceActive(deviceId)

			// Create database payload with scaled value
			deviceData := model.DeviceData{
				DeviceId:   deviceId,
				DeviceName: req.DeviceName,
				ValueData:  int(math.Round(req.ValueData * model.DeviceScale)),
			}

			// Send to Batcher
			svc.Add(deviceData)
		}
	}
}
