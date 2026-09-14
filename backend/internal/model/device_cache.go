package model

import "context"

type DeviceCacheService interface {
	GetDeviceInfoByName(ctx context.Context, deviceName string) (int, string, error)
	GetGroupInfoByName(ctx context.Context, groupName string) ([]int, string, error)
	InvalidateDevice(deviceName string)
	InvalidateGroup(groupName string)
	StartSweeper(ctx context.Context) // Replaces your background cache sweeper
}
