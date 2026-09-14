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
	deviceNameCache      sync.Map
	deviceGroupNameCache sync.Map
}

func NewCacheService(repo model.DeviceGatewayRepository) model.DeviceCacheService {
	return &cacheDeviceService{
		repo: repo,
	}
}

func (s *cacheDeviceService) GetDeviceInfoByName(ctx context.Context, deviceName string) (int, string, error) {
	// Step 1: Single Cache Lookup
	if item, ok := s.deviceNameCache.Load(deviceName); ok {
		cached := item.(cacheDeviceItem)
		if time.Now().Before(cached.expiresAt) {
			// Extract our struct
			info := cached.data.(cachedDeviceInfo)
			return info.DeviceId, info.Protocol, nil // ⚡ HIT
		}
		s.deviceNameCache.Delete(deviceName) // EXPIRED
	}

	// Step 2: Single Database Query Fallback
	dbId, protocol, err := s.repo.GetDeviceInfoByName(ctx, deviceName)
	if err != nil || dbId <= 0 {
		return 0, "", fmt.Errorf("device not found: %w", err)
	}

	// Step 3: Store them together in the single map
	s.deviceNameCache.Store(deviceName, cacheDeviceItem{
		data: cachedDeviceInfo{
			DeviceId: dbId,
			Protocol: protocol,
		},
		expiresAt: time.Now().Add(cacheTTL),
	})

	return dbId, protocol, nil
}

func (s *cacheDeviceService) GetGroupInfoByName(ctx context.Context, groupName string) ([]int, string, error) {
	// Step 1: Check in Cache
	if item, ok := s.deviceGroupNameCache.Load(groupName); ok {
		cached := item.(cacheDeviceItem)
		if time.Now().Before(cached.expiresAt) {
			// Extract our new struct
			info := cached.data.(cachedGroupInfo)
			return info.DeviceIds, info.Protocol, nil // ⚡ HIT
		}
		s.deviceGroupNameCache.Delete(groupName) // EXPIRED
	}

	// Step 2: Database Query Fallback
	// Your repo method already returns the protocol!
	groupDataList, err := s.repo.GetDeviceIdByGroupName(ctx, groupName)
	if err != nil || len(groupDataList) == 0 {
		return nil, "", fmt.Errorf("group not found or empty")
	}

	deviceIds := groupDataList[0].DeviceIds
	protocol := groupDataList[0].Protocol

	// Step 3: Store them together in the cache
	s.deviceGroupNameCache.Store(groupName, cacheDeviceItem{
		data: cachedGroupInfo{
			DeviceIds: deviceIds,
			Protocol:  protocol,
		},
		expiresAt: time.Now().Add(cacheTTL),
	})

	return deviceIds, protocol, nil
}

func (s *cacheDeviceService) InvalidateDevice(deviceName string) {
	s.deviceNameCache.Delete(deviceName)
}

func (s *cacheDeviceService) InvalidateGroup(groupName string) {
	s.deviceGroupNameCache.Delete(groupName)
}

// StartSweeper replaces the global sweeper you had in the listener package
func (s *cacheDeviceService) StartSweeper(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.cleanMap(&s.deviceNameCache, "deviceNameCache")
				s.cleanMap(&s.deviceGroupNameCache, "deviceGroupNameCache")
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
