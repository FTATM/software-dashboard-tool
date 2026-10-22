package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/FTATM/software-dashboard-tool/internal/model"
)

const cacheTTL = 5 * time.Minute
const cleanupInterval = 10 * time.Minute

type cachedDeviceInfo struct {
	DeviceId int
	Protocol string
}

type cachedGroupInfo struct {
	DeviceIds []int
	Protocol  string
}

type cacheDeviceItem struct {
	data      any
	expiresAt time.Time
}

type cacheDeviceService struct {
	repo                 model.DeviceGatewayRepository
	deviceNameCache      sync.Map // deviceName -> cachedDeviceInfo
	deviceGroupNameCache sync.Map // groupName -> cachedGroupInfo
	deviceIdCache        sync.Map // deviceId -> deviceName
	deviceGroupIdCache   sync.Map // groupId -> groupName
}

func NewCacheService(repo model.DeviceGatewayRepository) model.DeviceCacheService {
	return &cacheDeviceService{
		repo: repo,
	}
}

func (s *cacheDeviceService) GetDeviceInfoByName(ctx context.Context, deviceName string) (int, string, error) {
	if item, ok := s.deviceNameCache.Load(deviceName); ok {
		cached := item.(cacheDeviceItem)
		if time.Now().Before(cached.expiresAt) {
			info := cached.data.(cachedDeviceInfo)
			return info.DeviceId, info.Protocol, nil
		}
		s.deviceNameCache.Delete(deviceName)
	}

	dbId, protocol, err := s.repo.GetDeviceInfoByName(ctx, deviceName)
	if err != nil || dbId <= 0 {
		return 0, "", fmt.Errorf("device not found: %w", err)
	}

	exp := time.Now().Add(cacheTTL)
	s.deviceNameCache.Store(deviceName, cacheDeviceItem{
		data: cachedDeviceInfo{
			DeviceId: dbId,
			Protocol: protocol,
		},
		expiresAt: exp,
	})
	s.deviceIdCache.Store(dbId, cacheDeviceItem{
		data:      deviceName,
		expiresAt: exp,
	})

	return dbId, protocol, nil
}

func (s *cacheDeviceService) GetGroupInfoByName(ctx context.Context, groupName string) ([]int, string, error) {
	if item, ok := s.deviceGroupNameCache.Load(groupName); ok {
		cached := item.(cacheDeviceItem)
		if time.Now().Before(cached.expiresAt) {
			info := cached.data.(cachedGroupInfo)
			return info.DeviceIds, info.Protocol, nil
		}
		s.deviceGroupNameCache.Delete(groupName)
	}

	groupDataList, err := s.repo.GetDeviceIdByGroupName(ctx, groupName)
	if err != nil || len(groupDataList) == 0 {
		return nil, "", fmt.Errorf("group not found or empty")
	}

	deviceIds := groupDataList[0].DeviceIds
	protocol := groupDataList[0].Protocol
	groupId := groupDataList[0].GroupId

	exp := time.Now().Add(cacheTTL)
	s.deviceGroupNameCache.Store(groupName, cacheDeviceItem{
		data: cachedGroupInfo{
			DeviceIds: deviceIds,
			Protocol:  protocol,
		},
		expiresAt: exp,
	})
	if groupId > 0 {
		s.deviceGroupIdCache.Store(groupId, cacheDeviceItem{
			data:      groupName,
			expiresAt: exp,
		})
	}

	return deviceIds, protocol, nil
}

func (s *cacheDeviceService) GetDeviceNameById(ctx context.Context, deviceId int) (string, error) {
	if item, ok := s.deviceIdCache.Load(deviceId); ok {
		cached := item.(cacheDeviceItem)
		if time.Now().Before(cached.expiresAt) {
			return cached.data.(string), nil
		}
		s.deviceIdCache.Delete(deviceId)
	}

	name, err := s.repo.GetDeviceNameById(ctx, deviceId)
	if err != nil || name == "" {
		return "", fmt.Errorf("device not found for id %d: %w", deviceId, err)
	}

	s.deviceIdCache.Store(deviceId, cacheDeviceItem{
		data:      name,
		expiresAt: time.Now().Add(cacheTTL),
	})
	return name, nil
}

func (s *cacheDeviceService) GetGroupNameById(ctx context.Context, groupId int) (string, error) {
	if item, ok := s.deviceGroupIdCache.Load(groupId); ok {
		cached := item.(cacheDeviceItem)
		if time.Now().Before(cached.expiresAt) {
			return cached.data.(string), nil
		}
		s.deviceGroupIdCache.Delete(groupId)
	}

	name, err := s.repo.GetDeviceGroupNameById(ctx, groupId)
	if err != nil || name == "" {
		return "", fmt.Errorf("group not found for id %d: %w", groupId, err)
	}

	s.deviceGroupIdCache.Store(groupId, cacheDeviceItem{
		data:      name,
		expiresAt: time.Now().Add(cacheTTL),
	})
	return name, nil
}

func (s *cacheDeviceService) InvalidateDevice(deviceName string) {
	if item, ok := s.deviceNameCache.LoadAndDelete(deviceName); ok {
		if cached, ok := item.(cacheDeviceItem); ok {
			if info, ok := cached.data.(cachedDeviceInfo); ok {
				s.deviceIdCache.Delete(info.DeviceId)
				return
			}
		}
	}
	s.deviceIdCache.Range(func(key, value any) bool {
		if cached, ok := value.(cacheDeviceItem); ok {
			if name, ok := cached.data.(string); ok && name == deviceName {
				s.deviceIdCache.Delete(key)
			}
		}
		return true
	})
}

func (s *cacheDeviceService) InvalidateGroup(groupName string) {
	s.deviceGroupNameCache.Delete(groupName)
	s.deviceGroupIdCache.Range(func(key, value any) bool {
		if cached, ok := value.(cacheDeviceItem); ok {
			if name, ok := cached.data.(string); ok && name == groupName {
				s.deviceGroupIdCache.Delete(key)
			}
		}
		return true
	})
}

func (s *cacheDeviceService) StartSweeper(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.cleanMap(&s.deviceNameCache, "deviceNameCache")
				s.cleanMap(&s.deviceGroupNameCache, "deviceGroupNameCache")
				s.cleanMap(&s.deviceIdCache, "deviceIdCache")
				s.cleanMap(&s.deviceGroupIdCache, "deviceGroupIdCache")
			case <-ctx.Done():
				ticker.Stop()
				slog.Info("Shutting down cache sweeper...")
				return
			}
		}
	}()
}

func (s *cacheDeviceService) cleanMap(cache *sync.Map, cacheName string) {
	now := time.Now()
	expiredCount := 0

	cache.Range(func(key, value any) bool {
		cached := value.(cacheDeviceItem)
		if now.After(cached.expiresAt) {
			cache.Delete(key)
			expiredCount++
		}
		return true
	})

	if expiredCount > 0 {
		slog.Debug("Cache swept", slog.String("cache", cacheName), slog.Int("cleared", expiredCount))
	}
}
