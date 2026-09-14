package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const activeThreshold = 2 * time.Minute
const staleThreshold = 3 * time.Hour // How long before a device is considered permanently dead
const sweepInterval = 1 * time.Hour  // How often the sweeper checks memory

type sessionManagerService struct {
	repo model.DeviceGatewayRepository

	// Protocol Clients
	mqttClient mqtt.Client
	udpServer  *net.UDPConn

	// Thread-safe registries mapped by Device ID
	mu          sync.RWMutex
	tcpConns    map[int]net.Conn
	udpAddrs    map[int]*net.UDPAddr
	httpQueues  map[int][]model.DeviceCommand
	lastSeenMap map[int]time.Time

	deviceCacheSvc model.DeviceCacheService
}

// NewSessionManagerService satisfies the injection signature in app.go
func NewSessionManagerService(repo model.DeviceGatewayRepository, deviceCacheSvc model.DeviceCacheService) model.SessionManagerService {
	return &sessionManagerService{
		repo:           repo,
		tcpConns:       make(map[int]net.Conn),
		udpAddrs:       make(map[int]*net.UDPAddr),
		httpQueues:     make(map[int][]model.DeviceCommand),
		lastSeenMap:    make(map[int]time.Time),
		deviceCacheSvc: deviceCacheSvc,
	}
}

// --- SETUP PROTOCOLS ---

func (s *sessionManagerService) SetMQTTClient(client any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c, ok := client.(mqtt.Client); ok {
		s.mqttClient = c
		slog.Info("SessionManager: MQTT client registered")
	}
}

func (s *sessionManagerService) SetUDPServer(conn *net.UDPConn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.udpServer = conn
	slog.Info("SessionManager: UDP server registered")
}

// --- DEVICE REGISTRATIONS ---

func (s *sessionManagerService) RegisterTCP(deviceId int, conn net.Conn) {
	s.mu.Lock()
	var oldConn net.Conn

	// 1. ตรวจสอบว่ามี Connection เดิมอยู่หรือไม่ และไม่ใช่ Socket ตัวเดิม
	if existing, exists := s.tcpConns[deviceId]; exists && existing != conn {
		oldConn = existing
	}

	// 2. บันทึก Socket ตัวใหม่เข้าไปแทนที่
	s.tcpConns[deviceId] = conn
	s.mu.Unlock() // ปล่อย Mutex ทันทีเพื่อไม่ให้ระบบภาพรวมติดขัด

	// 3. หากมี Connection เก่า ให้ส่งข้อความแจ้งเตือนและปิดสายแบบ Asynchronous
	if oldConn != nil {
		go func(c net.Conn, id int) {
			defer c.Close()

			// ตั้งเวลาจำกัด (Timeout) ไว้ 1 วินาที ป้องกันกรณี Socket ค้าง
			_ = c.SetWriteDeadline(time.Now().Add(1 * time.Second))

			// ใช้ fmt.Appendf ตามคำแนะนำการ Optimize ของ Go
			msg := fmt.Appendf(nil, "DISCONNECTED: session replaced by new connection for device ID %d\n", id)

			if _, err := c.Write(msg); err != nil {
				slog.Debug("Failed to send disconnect reason to old TCP socket",
					slog.Int("deviceId", id),
					slog.String("error", err.Error()),
				)
			}
		}(oldConn, deviceId)
	}
}

func (s *sessionManagerService) UnregisterTCP(deviceId int, conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, exists := s.tcpConns[deviceId]; exists && existing == conn {
		delete(s.tcpConns, deviceId)
	}
}

func (s *sessionManagerService) RegisterUDP(deviceId int, addr *net.UDPAddr) {
	s.mu.Lock()

	// 1. ดึง IP/Port เดิมออกมาก่อน (ถ้ามี)
	oldAddr, exists := s.udpAddrs[deviceId]

	// 2. ดึง instance ของ UDP server ออกมาเพื่อใช้ส่งข้อมูล
	server := s.udpServer

	// 3. บันทึก IP/Port ใหม่ทับลงไป
	s.udpAddrs[deviceId] = addr

	s.mu.Unlock() // ปล่อย Lock ให้เร็วที่สุด

	// 4. หากเคยมี IP/Port เดิม และไม่ใช่ IP/Port ตัวเดียวกันกับปัจจุบัน ให้ส่งข้อความแจ้งเตือน
	if exists && oldAddr != nil && server != nil && oldAddr.String() != addr.String() {
		go func(target *net.UDPAddr, id int) {
			// สร้างข้อความด้วย fmt.Appendf เพื่อประสิทธิภาพ
			msg := fmt.Appendf(nil, "DISCONNECTED: session replaced by new UDP address for device ID %d\n", id)

			// ยิง UDP Packet ไปหา IP/Port เก่า
			if _, err := server.WriteToUDP(msg, target); err != nil {
				slog.Debug("Failed to send disconnect reason to old UDP address",
					slog.Int("deviceId", id),
					slog.String("error", err.Error()),
				)
			}
		}(oldAddr, deviceId)
	}
}

func (s *sessionManagerService) MarkDeviceActive(deviceId int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastSeenMap[deviceId] = time.Now()
}

func (s *sessionManagerService) IsDeviceOnline(deviceId int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lastSeen, exists := s.lastSeenMap[deviceId]
	if !exists {
		return false
	}

	// Consider online if seen within threshold, OR if it has an active TCP socket
	isRecent := time.Since(lastSeen) <= activeThreshold
	_, hasTCP := s.tcpConns[deviceId]

	return isRecent || hasTCP
}

// --- COMMAND ROUTING ---

func (s *sessionManagerService) RouteCommand(ctx context.Context, req *model.GatewayCommand) error {
	switch req.Protocol {
	case "MQTT":
		return s.routeMQTT(ctx, req)
	case "TCP":
		return s.routeTCP(ctx, req)
	case "UDP":
		return s.routeUDP(ctx, req)
	case "HTTP":
		return s.routeHTTP(ctx, req)
	default:
		return fmt.Errorf("unsupported protocol: %s", req.Protocol)
	}
}

func (s *sessionManagerService) routeMQTT(ctx context.Context, req *model.GatewayCommand) error {
	s.mu.RLock()
	client := s.mqttClient
	s.mu.RUnlock()

	if client == nil {
		return fmt.Errorf("MQTT client not initialized")
	}

	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return err
	}

	var topic string
	if req.GroupId > 0 {
		groupName, err := s.repo.GetDeviceGroupNameById(ctx, req.GroupId)
		if err != nil || groupName == "" {
			return fmt.Errorf("failed to fetch group name for id %d: %w", req.GroupId, err)
		}
		topic = fmt.Sprintf("device-group/%s/cmd", groupName)
	} else if req.DeviceId > 0 {
		deviceName, err := s.repo.GetDeviceNameById(ctx, req.DeviceId)
		if err != nil || deviceName == "" {
			return fmt.Errorf("failed to fetch device name for id %d: %w", req.DeviceId, err)
		}
		topic = fmt.Sprintf("device/%s/cmd", deviceName)
	} else {
		return fmt.Errorf("missing target ID for MQTT command")
	}

	token := client.Publish(topic, 1, false, payloadBytes)
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	slog.InfoContext(ctx, "Routed command via MQTT", slog.String("topic", topic))
	return nil
}

func (s *sessionManagerService) routeTCP(ctx context.Context, req *model.GatewayCommand) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// ⚡ 1. ใช้ Map เพื่อกรองการเชื่อมต่อที่ซ้ำกัน (Deduplication สำหรับ Gateway)
	targetConns := make(map[net.Conn]bool)
	var devicesToNotify []int

	// ⚡ 2. ค้นหาเป้าหมาย (รองรับทั้ง Group และ Single)
	if req.GroupId > 0 {
		groupName, err := s.repo.GetDeviceGroupNameById(ctx, req.GroupId)
		if err == nil && groupName != "" {
			// ใช้ฟังก์ชันที่คุณสร้างไว้ในรอบที่แล้ว
			deviceIds, _, _ := s.deviceCacheSvc.GetGroupInfoByName(ctx, groupName)
			devicesToNotify = append(devicesToNotify, deviceIds...)
		}
	} else if req.DeviceId > 0 {
		devicesToNotify = append(devicesToNotify, req.DeviceId)
	} else {
		// Fallback กลับไปอ่านจาก Payload หากไม่ได้ระบุ ID มา
		for _, cmd := range req.Payload {
			id, _, _ := s.deviceCacheSvc.GetDeviceInfoByName(ctx, cmd.DeviceName)
			if id > 0 {
				devicesToNotify = append(devicesToNotify, id)
			}
		}
	}

	// ⚡ 3. รวบรวม Socket (net.Conn) ที่ใช้งานอยู่จริงของอุปกรณ์ทั้งหมด
	for _, id := range devicesToNotify {
		if conn, exists := s.tcpConns[id]; exists {
			targetConns[conn] = true
		}
	}

	if len(targetConns) == 0 {
		slog.Warn("No active TCP connections found for routing")
		return nil
	}

	// แปลงข้อมูลคำสั่งเป็น JSON
	cmdBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return err
	}
	cmdBytes = append(cmdBytes, '\n') // ปิดท้ายด้วย Newline สำหรับ TCP

	// ⚡ 4. ส่งคำสั่งไปยังแต่ละ Socket (Gateway จะได้รับแค่ 1 ครั้งแม้จะดูแลอุปกรณ์ 10 ตัวก็ตาม)
	for conn := range targetConns {
		if _, err := conn.Write(cmdBytes); err != nil {
			slog.Error("Failed to write to TCP connection", slog.String("err", err.Error()))
		}
	}

	slog.InfoContext(ctx, "Routed command via TCP", slog.Int("connections_sent", len(targetConns)))
	return nil
}

func (s *sessionManagerService) routeUDP(ctx context.Context, req *model.GatewayCommand) error {
	s.mu.RLock()
	server := s.udpServer
	s.mu.RUnlock()

	if server == nil {
		return fmt.Errorf("UDP server not initialized")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 1. ใช้ Map เพื่อกรอง IP/Port ที่ซ้ำกัน (Deduplication สำหรับ UDP Gateway)
	targetAddrs := make(map[string]*net.UDPAddr)
	var devicesToNotify []int

	// 2. ค้นหาเป้าหมาย (รองรับทั้ง Group และ Single เหมือน TCP)
	if req.GroupId > 0 {
		groupName, err := s.repo.GetDeviceGroupNameById(ctx, req.GroupId)
		if err == nil && groupName != "" {
			deviceIds, _, _ := s.deviceCacheSvc.GetGroupInfoByName(ctx, groupName)
			devicesToNotify = append(devicesToNotify, deviceIds...)
		}
	} else if req.DeviceId > 0 {
		devicesToNotify = append(devicesToNotify, req.DeviceId)
	} else {
		// Fallback กลับไปอ่านจาก Payload
		for _, cmd := range req.Payload {
			id, _, _ := s.deviceCacheSvc.GetDeviceInfoByName(ctx, cmd.DeviceName)
			if id > 0 {
				devicesToNotify = append(devicesToNotify, id)
			}
		}
	}

	// 3. รวบรวมที่อยู่ (UDPAddr) ที่ใช้งานอยู่จริงของอุปกรณ์ทั้งหมด
	for _, id := range devicesToNotify {
		if addr, exists := s.udpAddrs[id]; exists {
			// ใช้ addr.String() เป็น Key เพื่อกรอง IP:Port ที่ซ้ำกัน
			targetAddrs[addr.String()] = addr
		}
	}

	if len(targetAddrs) == 0 {
		slog.Warn("No active UDP addresses found for routing")
		return nil
	}

	cmdBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return err
	}

	// 4. ยิงคำสั่งไปยังแต่ละ IP/Port (Gateway จะได้รับแค่ 1 ครั้ง)
	for _, addr := range targetAddrs {
		if _, err := server.WriteToUDP(cmdBytes, addr); err != nil {
			slog.Error("Failed to write to UDP", slog.String("err", err.Error()))
		}
	}

	slog.InfoContext(ctx, "Routed command via UDP", slog.Int("addresses_sent", len(targetAddrs)))
	return nil
}

func (s *sessionManagerService) routeHTTP(ctx context.Context, req *model.GatewayCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, cmd := range req.Payload {
		id, _, err := s.deviceCacheSvc.GetDeviceInfoByName(ctx, cmd.DeviceName)
		if err != nil || id <= 0 {
			continue
		}
		s.httpQueues[id] = append(s.httpQueues[id], cmd)
	}
	return nil
}

// --- HTTP POLLING QUEUE ---

func (s *sessionManagerService) PopHTTPCommand(deviceId int) ([]model.DeviceCommand, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	commands, exists := s.httpQueues[deviceId]
	if !exists || len(commands) == 0 {
		return nil, false
	}

	// Empty the queue after popping
	delete(s.httpQueues, deviceId)
	return commands, true
}

func (s *sessionManagerService) StartSweeper(ctx context.Context) {
	ticker := time.NewTicker(sweepInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.sweepStaleConnections()
			case <-ctx.Done():
				ticker.Stop()
				slog.Info("Shutting down session manager sweeper...")
				return
			}
		}
	}()
}

func (s *sessionManagerService) sweepStaleConnections() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	clearedCount := 0

	// Iterate through the lastSeenMap to find devices that haven't reported in
	for deviceId, lastSeen := range s.lastSeenMap {
		if now.Sub(lastSeen) > staleThreshold {
			// Device is stale, clear all its memory footprints
			delete(s.lastSeenMap, deviceId)
			delete(s.httpQueues, deviceId)
			delete(s.udpAddrs, deviceId)

			// If a TCP connection somehow got orphaned without triggering UnregisterTCP, forcefully close and delete it
			if conn, exists := s.tcpConns[deviceId]; exists {
				conn.Close()
				delete(s.tcpConns, deviceId)
			}

			clearedCount++
		}
	}

	if clearedCount > 0 {
		slog.Debug("Session manager swept dead connections", slog.Int("cleared", clearedCount))
	}
}
