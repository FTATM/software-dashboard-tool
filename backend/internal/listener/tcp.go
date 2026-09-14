package listener

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net"
	"time"

	"github.com/FTATM/software-dashboard-tool/internal/model"
)

func StartTCPServer(ctx context.Context, port string, svc model.DeviceGatewayService, sessionSvc model.SessionManagerService, deviceCacheSvc model.DeviceCacheService) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to start TCP listener", slog.String("error", err.Error()))
		return
	}
	slog.InfoContext(ctx, "TCP Listener started", slog.String("port", port))

	// Wait for shutdown signal to close the listener
	go func() {
		<-ctx.Done()
		slog.Info("Shutting down TCP listener...")
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return // Context canceled, safe exit
			}
			slog.WarnContext(ctx, "Failed to accept TCP connection", slog.String("error", err.Error()))
			continue
		}

		// Handle each device in its own goroutine so they don't block each other
		go handleTCPConnection(ctx, conn, svc, sessionSvc, deviceCacheSvc)
	}
}

func handleTCPConnection(ctx context.Context, conn net.Conn, svc model.DeviceGatewayService, sessionSvc model.SessionManagerService, deviceCacheSvc model.DeviceCacheService) {
	remoteAddr := conn.RemoteAddr().String()
	slog.InfoContext(ctx, "Device connected via TCP", slog.String("ip", remoteAddr))

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(3 * time.Minute)
	}

	// ⚡ 1. ใช้ Slice และ Map เพื่อติดตามอุปกรณ์หลายตัวบน Connection เดียว (Gateway Mode)
	var connectedDeviceIDs []int
	registeredDevices := make(map[int]bool)

	defer func() {
		conn.Close()
		// ยกเลิกการลงทะเบียนอุปกรณ์ทั้งหมดที่ผูกกับ Connection นี้
		for _, id := range connectedDeviceIDs {
			sessionSvc.UnregisterTCP(id, conn)
		}
		slog.InfoContext(ctx, "Device/Gateway disconnected from TCP", slog.String("ip", remoteAddr))
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		rawLine := bytes.TrimSpace(scanner.Bytes())
		if len(rawLine) == 0 {
			continue
		}

		var incomingData []model.DeviceDataPayloadReq
		switch rawLine[0] {
		case '[':
			if err := json.Unmarshal(rawLine, &incomingData); err != nil {
				continue
			}
		case '{':
			var single model.DeviceDataPayloadReq
			if err := json.Unmarshal(rawLine, &single); err != nil {
				continue
			}
			incomingData = append(incomingData, single)
		default:
			continue
		}

		for _, req := range incomingData {
			if req.DeviceName == "" {
				continue
			}

			deviceId, protocol, err := deviceCacheSvc.GetDeviceInfoByName(ctx, req.DeviceName)
			if err != nil || deviceId <= 0 {
				slog.Warn("Unknown or empty device name received", slog.String("groupName", req.DeviceName))
				continue
			}

			if protocol != "TCP" {
				slog.Warn("Device protocol mismatch", slog.String("DeviceName", req.DeviceName), slog.String("expected", "TCP"), slog.String("got", protocol))
				continue
			}

			// ⚡ 2. หากพบ Device ใหม่ที่เพิ่งส่งข้อมูลมาทาง Socket นี้ ให้ลงทะเบียนผูกกับ Socket ทันที
			if !registeredDevices[deviceId] {
				registeredDevices[deviceId] = true
				connectedDeviceIDs = append(connectedDeviceIDs, deviceId)
				sessionSvc.RegisterTCP(deviceId, conn)

				// ⚡ NEW: Send ACK back to the device/gateway
				ackMsg := fmt.Appendf(nil, "TCP %s connected!\n", req.DeviceName)
				if _, err := conn.Write(ackMsg); err != nil {
					slog.WarnContext(ctx, "Failed to send ACK to TCP device",
						slog.String("device", req.DeviceName),
						slog.String("error", err.Error()),
					)
				}
			}

			sessionSvc.MarkDeviceActive(deviceId)

			deviceData := model.DeviceData{
				DeviceId:   deviceId,
				DeviceName: req.DeviceName,
				ValueData:  int(math.Round(req.ValueData * model.DeviceScale)),
			}
			svc.Add(deviceData)
		}
	}

	if err := scanner.Err(); err != nil {
		slog.ErrorContext(ctx, "TCP stream read error", slog.String("error", err.Error()))
	}
}
