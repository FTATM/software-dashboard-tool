package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type deviceService struct {
	txManager    model.TransactionManager
	prefixError  string
	deviceRepo   model.DeviceRepository
	auditLogRepo model.AuditLogRepository
	clientsMap   map[string][]chan model.ChartData
	mutex        sync.RWMutex
}

func NewDeviceService(txManager model.TransactionManager, dr model.DeviceRepository, alog model.AuditLogRepository) model.DeviceService {
	return &deviceService{
		txManager:    txManager,
		prefixError:  "deviceService",
		deviceRepo:   dr,
		auditLogRepo: alog,
		clientsMap:   make(map[string][]chan model.ChartData),
		mutex:        sync.RWMutex{},
	}
}

func (s *deviceService) GetAllDeviceDetail(ctx context.Context) ([]model.DeviceDetail, error) {
	const fname = "GetAllDeviceDetail"
	devices, err := s.deviceRepo.GetAll(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	deviceDetails := make([]model.DeviceDetail, 0, len(devices))

	for _, d := range devices {
		detail := model.DeviceDetail{
			DeviceId:   d.DeviceId,
			DeviceName: d.DeviceName,
			Protocol:   d.Protocol,
			ValueData:  d.ValueData,
			Active:     d.Active,
			LastSeenAt: d.LastSeenAt,
		}
		deviceDetails = append(deviceDetails, detail)
	}

	return deviceDetails, nil
}

func (s *deviceService) CreateDevice(ctx context.Context, createDevice []model.DeviceCreate, authUserId int) error {
	const fname = "CreateDevice"
	var err error
	if len(createDevice) == 0 {
		return nil
	}

	devices := make([]model.Device, 0, len(createDevice))
	for _, createReq := range createDevice {
		device := model.Device{
			DeviceName: createReq.DeviceName,
			Protocol:   createReq.Protocol,
			Active:     createReq.Active,
		}
		devices = append(devices, device)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	if err = s.deviceRepo.Create(tx.Context(), devices); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, len(devices))
	for _, d := range devices {

		newData, err := model.StructToDynamicJSON(d)
		if err != nil {
			return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}
		audit := model.AuditLog{
			EntityType: "device",
			EntityId:   strconv.Itoa(d.DeviceId),
			MenuType:   "Device",
			Action:     model.CreateAction,
			ChangedBy:  authUserId,
			OldData:    nil,
			NewData:    newData,
		}
		auditlogs = append(auditlogs, audit)
	}

	if err = s.auditLogRepo.Create(tx.Context(), auditlogs); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *deviceService) UpdateDevice(ctx context.Context, updateDevice *model.DeviceUpdate, authUserId int) error {
	const fname = "UpdateDevice"
	var err error

	device := model.Device{
		DeviceName: updateDevice.DeviceName,
		DeviceId:   updateDevice.DeviceId,
		Active:     updateDevice.Active,
		Protocol:   updateDevice.Protocol,
	}

	oldDevice, err := s.deviceRepo.GetById(ctx, device.DeviceId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	// check same then return
	if oldDevice.IsSame(device) {
		return nil
	}

	updateDevice.OldName = oldDevice.DeviceName

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	if err = s.deviceRepo.Update(tx.Context(), &device); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, 1)
	oldData, err := model.StructToDynamicJSON(oldDevice)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	newData, err := model.StructToDynamicJSON(device)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "device",
		EntityId:   strconv.Itoa(device.DeviceId),
		MenuType:   "Device",
		Action:     model.UpdateAction,
		ChangedBy:  authUserId,
		OldData:    oldData,
		NewData:    newData,
	}
	auditlogs = append(auditlogs, audit)

	if err = s.auditLogRepo.Create(tx.Context(), auditlogs); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *deviceService) DeleteDevice(ctx context.Context, deviceId, authUserId int) (*model.Device, error) {
	const fname = "DeleteDevice"
	var err error

	device, err := s.deviceRepo.GetById(ctx, deviceId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	if err = s.deviceRepo.Delete(tx.Context(), deviceId); err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, 1)
	oldData, err := model.StructToDynamicJSON(device)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	audit := model.AuditLog{
		EntityType: "device",
		EntityId:   strconv.Itoa(deviceId),
		MenuType:   "Device",
		Action:     model.DeleteAction,
		ChangedBy:  authUserId,
		OldData:    oldData,
		NewData:    nil,
	}
	auditlogs = append(auditlogs, audit)

	if err = s.auditLogRepo.Create(tx.Context(), auditlogs); err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return device, nil
}

func (s *deviceService) GetProtocolType(ctx context.Context) ([]string, error) {
	const fname = "GetProtocolType"
	protocols, err := s.deviceRepo.GetProtocolType(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return protocols, nil
}

func (s *deviceService) AddClient(deviceId string, clientChan chan model.ChartData) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.clientsMap[deviceId] = append(s.clientsMap[deviceId], clientChan)
}

func (s *deviceService) RemoveClient(deviceId string, clientChan chan model.ChartData) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Find and remove the closed channel from the slice
	clients := s.clientsMap[deviceId]
	for i, ch := range clients {
		if ch == clientChan {
			s.clientsMap[deviceId] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
}

// Update StartPublic to use the struct's state
func (s *deviceService) StartPublic(ctx context.Context) {
	// ⚡ 1. Catch Nil Contexts immediately
	if ctx == nil {
		slog.Error("CRITICAL: StartPublic received a NIL context! Goroutine cannot start.")
		return
	}

	slog.Info("StartPublic worker successfully started in the background!")

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("🛑 StartPublic context cancelled. Shutting down worker.")
			return

		case <-ticker.C:
			// ⚡ 2. Prove the ticker is firing every 3 seconds
			slog.Debug("⏱️ StartPublic Ticker fired!")

			s.mutex.RLock()
			activeDeviceIds := make(map[int]bool)
			for requestedIDs, clients := range s.clientsMap {
				if len(clients) == 0 {
					continue
				}
				for idStr := range strings.SplitSeq(requestedIDs, ",") {
					if idInt, err := strconv.Atoi(idStr); err == nil {
						activeDeviceIds[idInt] = true
					}
				}
			}
			clientCount := len(s.clientsMap)
			s.mutex.RUnlock()

			// ⚡ 3. Prove how many clients the server thinks are currently connected
			slog.Debug("📊 StartPublic Status",
				slog.Int("connected_clients", clientCount),
				slog.Int("unique_devices_needed", len(activeDeviceIds)),
			)

			// If no clients are connected, skip the database call
			if len(activeDeviceIds) == 0 {
				continue
			}

			// ⚡ 4. Prove we reached the database phase
			slog.Debug("🗄️ StartPublic querying database for clients...")

			masterDataMap := make(map[int]model.ChartDeviceData)
			for deviceId := range activeDeviceIds {
				chartDeviceData, err := s.deviceRepo.GetByIdChartDeviceData(ctx, deviceId)
				if err != nil {
					slog.Error("Failed to fetch chart data for device",
						slog.Int("deviceId", deviceId),
						slog.String("error", err.Error()))
					continue
				}

				// ⚡ REVERSE SCALING: Divide by 100.0 to convert the DB integer (e.g., 2456)
				// back into the real decimal (24.56) for the Vue frontend.
				chartDeviceData.ValueData = chartDeviceData.ValueData / float64(model.DeviceScale)

				masterDataMap[deviceId] = chartDeviceData
			}

			s.mutex.RLock()
			for requestedIds, clients := range s.clientsMap {
				if len(clients) == 0 {
					continue
				}

				clientPayloadMap := make(map[int]model.ChartDeviceData)
				for idStr := range strings.SplitSeq(requestedIds, ",") {
					if idInt, err := strconv.Atoi(idStr); err == nil {
						clientPayloadMap[idInt] = masterDataMap[idInt]
					}
				}

				payload := model.ChartData{
					DeviceData: clientPayloadMap,
				}

				for _, ch := range clients {
					select {
					case ch <- payload:
						slog.Debug("Data successfully pushed to client channel!")
					default:
						slog.Warn("Client channel full, dropping packet")
					}
				}
			}
			s.mutex.RUnlock()
		}
	}
}

func (s *deviceService) GetAllDeviceName(ctx context.Context) ([]model.DeviceDetail, error) {
	const fname = "GetAllDeviceName"
	devices, err := s.deviceRepo.GetAllName(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	deviceDetails := make([]model.DeviceDetail, 0, len(devices))

	for _, d := range devices {
		detail := model.DeviceDetail{
			DeviceId:   d.DeviceId,
			DeviceName: d.DeviceName,
			Active:     d.Active,
		}
		deviceDetails = append(deviceDetails, detail)
	}

	return deviceDetails, nil
}

func (s *deviceService) GetChartHistory(ctx context.Context, deviceIds []int, maxPoints int, fromTime, toTime time.Time) (map[int][][2]float64, error) {
	const fname = "GetChartHistory"
	count, err := s.deviceRepo.CountData(ctx, deviceIds, fromTime, toTime)
	if err != nil {
		return nil, err
	}

	var logs []model.DeviceDataLog

	// 2. ⚡ DYNAMIC DECISION LOGIC ⚡
	if count <= maxPoints {
		// Scenario A: Safe to send raw data!
		// The user zoomed in, or the device hasn't sent many points yet.
		logs, err = s.deviceRepo.GetRawData(ctx, deviceIds, fromTime, toTime, maxPoints)
		if err != nil {
			return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}
	} else {
		// Scenario B: Too much data!
		// We must aggregate it so we don't crash the frontend.
		totalDuration := toTime.Sub(fromTime)

		// Calculate the dynamic bucket size
		bucketDuration := max(totalDuration/time.Duration(maxPoints), time.Second)
		bucketInterval := fmt.Sprintf("%f seconds", bucketDuration.Seconds())

		logs, err = s.deviceRepo.GetAggregatedData(ctx, deviceIds, fromTime, toTime, bucketInterval)
		if err != nil {
			return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}
	}

	// 3. Map the chosen data perfectly for ECharts
	historyData := make(map[int][][2]float64)

	for _, log := range logs {
		tsMillis := float64(log.ReceivedAt.UnixMilli())
		valData := float64(log.ValueData) / float64(model.DeviceScale)

		historyData[log.DeviceId] = append(historyData[log.DeviceId], [2]float64{tsMillis, valData})
	}

	return historyData, nil
}

func (s *deviceService) GetDeviceGroupDetail(ctx context.Context) ([]model.DeviceGroupDetail, error) {
	const fname = "GetDeviceGroupDetail"
	deviceGroups, err := s.deviceRepo.GetAllDeviceGroup(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	deviceGroupId := make([]int, 0, len(deviceGroups))
	for _, g := range deviceGroups {
		deviceGroupId = append(deviceGroupId, g.GroupId)
	}

	deviceGroupMap, err := s.deviceRepo.GetDeviceIdByDeviceGroupIds(ctx, deviceGroupId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	deviceGroupDetails := make([]model.DeviceGroupDetail, 0, len(deviceGroups))

	for _, d := range deviceGroups {
		detail := model.DeviceGroupDetail{
			DeviceGroup: d,
			DeviceIds:   deviceGroupMap[d.GroupId],
		}
		deviceGroupDetails = append(deviceGroupDetails, detail)
	}

	return deviceGroupDetails, nil
}

func (s *deviceService) CreateDeviceGroup(ctx context.Context, createDeviceG model.CreateDeviceGroup, authUserId int) error {
	const fname = "CreateDeviceGroup"
	var err error

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	deviceGroup := model.DeviceGroup{
		GroupName:   createDeviceG.GroupName,
		Description: createDeviceG.Description,
		Protocol:    createDeviceG.Protocol,
	}

	if err = s.deviceRepo.CreateGroup(tx.Context(), &deviceGroup); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	deviceIds := make([]int, 0, len(createDeviceG.DeviceIds))
	for _, id := range createDeviceG.DeviceIds {
		deviceIds = append(deviceIds, id)
	}

	if err = s.deviceRepo.CreateGroupMap(tx.Context(), deviceGroup.GroupId, deviceIds); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	DeviceGroupDetail := model.DeviceGroupDetail{
		DeviceGroup: deviceGroup,
		DeviceIds:   createDeviceG.DeviceIds,
	}

	newData, err := model.StructToDynamicJSON(DeviceGroupDetail)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "device_group",
		EntityId:   strconv.Itoa(deviceGroup.GroupId),
		MenuType:   "Device Group",
		Action:     model.CreateAction,
		ChangedBy:  authUserId,
		OldData:    nil,
		NewData:    newData,
	}

	if err = s.auditLogRepo.Create(tx.Context(), []model.AuditLog{audit}); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *deviceService) UpdateDeviceGroup(ctx context.Context, updateDeviceG *model.UpdateDeviceGroup, authUserId int) error {
	const fname = "UpdateDeviceGroup"
	var err error

	deviceGroup := model.DeviceGroup{
		GroupId:     updateDeviceG.GroupId,
		GroupName:   updateDeviceG.GroupName,
		Description: updateDeviceG.Description,
		Protocol:    updateDeviceG.Protocol,
	}

	oldDeviceGroup, err := s.deviceRepo.GetDeviceGroupById(ctx, deviceGroup.GroupId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	oldDeviceGroupMap, err := s.deviceRepo.GetDeviceIdByDeviceGroupIds(ctx, []int{deviceGroup.GroupId})
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	// 1. Build lookup sets
	oldDeviceIdSet := make(map[int]struct{}, len(oldDeviceGroupMap))
	for _, deviceIds := range oldDeviceGroupMap {
		for _, id := range deviceIds {
			oldDeviceIdSet[id] = struct{}{}
		}
	}

	newDeviceIdSet := make(map[int]struct{}, len(updateDeviceG.DeviceIds))
	for _, id := range updateDeviceG.DeviceIds {
		newDeviceIdSet[id] = struct{}{}
	}

	// 2. Collect IDs to ADD (in new, but not in old)
	var toAddMap []int
	for _, id := range updateDeviceG.DeviceIds {
		if _, exists := oldDeviceIdSet[id]; !exists {
			toAddMap = append(toAddMap, id)
		}
	}

	// 3. Collect IDs to DELETE (in old, but not in new)
	var toDeleteMap []int
	for _, deviceIds := range oldDeviceGroupMap {
		for _, id := range deviceIds {
			if _, exists := newDeviceIdSet[id]; !exists {
				toDeleteMap = append(toDeleteMap, id)
			}
		}
	}

	if oldDeviceGroup.IsSame(deviceGroup) && len(toAddMap) == 0 && len(toDeleteMap) == 0 {
		return nil
	}

	updateDeviceG.OldName = oldDeviceGroup.GroupName

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	if err = s.deviceRepo.UpdateGroup(tx.Context(), &deviceGroup); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err = s.deviceRepo.CreateGroupMap(tx.Context(), deviceGroup.GroupId, toAddMap); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err = s.deviceRepo.DeleteGroupMap(tx.Context(), deviceGroup.GroupId, toDeleteMap); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	newDeviceGroupDetail := model.DeviceGroupDetail{
		DeviceGroup: deviceGroup,
		DeviceIds:   updateDeviceG.DeviceIds,
	}

	oldDeviceGroupDetail := model.DeviceGroupDetail{
		DeviceGroup: *oldDeviceGroup,
		DeviceIds:   oldDeviceGroupMap[deviceGroup.GroupId],
	}

	newData, err := model.StructToDynamicJSON(newDeviceGroupDetail)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	oldData, err := model.StructToDynamicJSON(oldDeviceGroupDetail)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	audit := model.AuditLog{
		EntityType: "device_group",
		EntityId:   strconv.Itoa(deviceGroup.GroupId),
		MenuType:   "Device Group",
		Action:     model.UpdateAction,
		ChangedBy:  authUserId,
		OldData:    oldData,
		NewData:    newData,
	}

	if err = s.auditLogRepo.Create(tx.Context(), []model.AuditLog{audit}); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *deviceService) DeleteDeviceGroup(ctx context.Context, deviceGroupId, authUserId int) (*model.DeviceGroup, error) {
	const fname = "DeleteDeviceGroup"
	var err error

	deviceGroup, err := s.deviceRepo.GetDeviceGroupById(ctx, deviceGroupId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	deviceGroupMaps, err := s.deviceRepo.GetDeviceIdByDeviceGroupIds(ctx, []int{deviceGroup.GroupId})
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	if err = s.deviceRepo.DeleteGroup(tx.Context(), deviceGroupId); err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	deviceGroupDetail := model.DeviceGroupDetail{
		DeviceGroup: *deviceGroup,
		DeviceIds:   deviceGroupMaps[deviceGroupId],
	}

	oldData, err := model.StructToDynamicJSON(deviceGroupDetail)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	audit := model.AuditLog{
		EntityType: "device_group",
		EntityId:   strconv.Itoa(deviceGroupId),
		MenuType:   "Device Group",
		Action:     model.DeleteAction,
		ChangedBy:  authUserId,
		OldData:    oldData,
		NewData:    nil,
	}

	if err = s.auditLogRepo.Create(tx.Context(), []model.AuditLog{audit}); err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return deviceGroup, nil
}

func (s *deviceService) GetDeviceForCommandByIds(ctx context.Context, deviceIds []int) ([]model.CommandDeviceInfo, error) {
	const fname = "GetDeviceForCommandByIds"

	devices, err := s.deviceRepo.GetDeviceForCommandByIds(ctx, deviceIds)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return devices, nil
}
