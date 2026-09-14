package model

import (
	"context"
	"strconv"
	"time"
)

type DeviceDataPayloadReq struct {
	DeviceName string  `json:"deviceName,omitempty"`
	ValueData  float64 `json:"valueData"`
}

type DeviceData struct {
	DeviceId   int       `json:"deviceId" db:"device_id"`
	DeviceName string    `json:"deviceName" db:"-"`
	ValueData  int       `json:"valueData" db:"value_data"`
	ReceivedAt time.Time `json:"-"`
}

type GatewayCommand struct {
	DeviceId int             `json:"deviceId"`
	GroupId  int             `json:"groupId"`
	Payload  []DeviceCommand `json:"payload"`
	Protocol string          `json:"protocol"`
}

type DeviceCommand struct {
	DeviceName string `json:"deviceName"`
	Cmd        string `json:"cmd"`
}

type DeviceGatewayRepository interface {
	BulkUpsertDeviceData(ctx context.Context, data []DeviceData) error
	UpdateLastSeen(ctx context.Context, deviceId int) error
	GetDeviceInfoByName(ctx context.Context, deviceName string) (int, string, error)
	GetDeviceIdByGroupName(ctx context.Context, deviceName string) ([]DeviceGroupData, error)
	GetDeviceNameById(ctx context.Context, deviceId int) (string, error)
	GetDeviceGroupNameById(ctx context.Context, groupId int) (string, error)
}

type DeviceGatewayService interface {
	Add(data DeviceData)
	Start(ctx context.Context) error
	Stop()
	UpdateDeviceLastSeen(ctx context.Context, deviceId int) error
}

type DeviceGatewayClient interface {
	GetDeviceStatus(ctx context.Context, deviceId int) (bool, error)
	ExecuteCommand(ctx context.Context, req GatewayCommand) error
	InvalidateDeviceCache(ctx context.Context, oldDeviceName string) error
	InvalidateGroupCache(ctx context.Context, oldGroupName string) error
}

func BuildGatewayCommands(devices []CommandDeviceInfo, actionPayload TaskActionPayload, isGroupTarget bool) []GatewayCommand {
	groupedCommands := make(map[int]GatewayCommand)
	var finalCommands []GatewayCommand

	for _, dev := range devices {
		devIDStr := strconv.Itoa(dev.DeviceId)

		cmdStr := actionPayload.Command
		if override, exists := actionPayload.DeviceOverrides[devIDStr]; exists && override != "" {
			cmdStr = override
		}

		payloadItem := DeviceCommand{
			DeviceName: dev.DeviceName,
			Cmd:        cmdStr,
		}

		hasGatewayProtocol := dev.GroupProtocol != nil && *dev.GroupProtocol != "" && *dev.GroupProtocol != "none"

		if isGroupTarget && dev.GroupId != nil && *dev.GroupId > 0 && hasGatewayProtocol {
			grpID := *dev.GroupId
			if existing, ok := groupedCommands[grpID]; ok {
				existing.Payload = append(existing.Payload, payloadItem)
				groupedCommands[grpID] = existing
			} else {
				groupedCommands[grpID] = GatewayCommand{
					DeviceId: 0,
					GroupId:  grpID,
					Protocol: *dev.GroupProtocol,
					Payload:  []DeviceCommand{payloadItem},
				}
			}
		} else {
			protocol := "none"
			if dev.Protocol != nil && *dev.Protocol != "" {
				protocol = *dev.Protocol
			}
			finalCommands = append(finalCommands, GatewayCommand{
				DeviceId: dev.DeviceId,
				GroupId:  0,
				Protocol: protocol,
				Payload:  []DeviceCommand{payloadItem},
			})
		}
	}

	for _, cmd := range groupedCommands {
		finalCommands = append(finalCommands, cmd)
	}

	return finalCommands
}
