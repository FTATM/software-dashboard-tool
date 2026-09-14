package model

import (
	"context"
	"net"
)

// SessionManagerService defines the contract for managing active connections & command routing
type SessionManagerService interface {
	SetMQTTClient(client any)
	SetUDPServer(conn *net.UDPConn)
	RegisterTCP(deviceId int, conn net.Conn)
	UnregisterTCP(deviceId int, conn net.Conn)
	RegisterUDP(deviceId int, addr *net.UDPAddr)
	RouteCommand(ctx context.Context, req *GatewayCommand) error
	PopHTTPCommand(deviceId int) ([]DeviceCommand, bool)
	MarkDeviceActive(deviceId int)
	IsDeviceOnline(deviceId int) bool
	StartSweeper(ctx context.Context)
}
