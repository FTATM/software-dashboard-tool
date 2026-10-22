package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type scheduleEngineService struct {
	scheduleEngineRepo     model.ScheduleEngineRepository
	prefixError            string
	engine                 gocron.Scheduler
	deviceRepo             model.DeviceRepository
	gatewayClient          model.DeviceGatewayClient
	jobRegistry            sync.Map
	maintenanceCache       sync.Map
	widgetRepo             model.WidgetRepository
	s3Client               model.S3Client
	auditLogRepo           model.AuditLogRepository
	notifRepo              model.NotificationRepository
	notifClient            model.NotificationClient
	maintenanceNotifRepeat int
}

func NewSchedulerEngineService(
	ser model.ScheduleEngineRepository,
	dr model.DeviceRepository,
	gc model.DeviceGatewayClient,
	widgetRepo model.WidgetRepository,
	s3Client model.S3Client,
	auditLogRepo model.AuditLogRepository,
	notifRepo model.NotificationRepository,
	notifClient model.NotificationClient,
	maintenanceNotifRepeat int,
) (model.ScheduleEngineService, error) {
	engine, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}
	return &scheduleEngineService{
		scheduleEngineRepo:     ser,
		deviceRepo:             dr,
		gatewayClient:          gc,
		engine:                 engine,
		prefixError:            "scheduleEngineService",
		widgetRepo:             widgetRepo,
		s3Client:               s3Client,
		auditLogRepo:           auditLogRepo,
		notifRepo:              notifRepo,
		notifClient:            notifClient,
		maintenanceNotifRepeat: maintenanceNotifRepeat,
	}, nil
}

func (s *scheduleEngineService) SyncJob(ctx context.Context, id string) error {
	const fname = "SyncJob"
	// 1. Fetch the newly saved data from Postgres
	sched, err := s.scheduleEngineRepo.GetById(ctx, id)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}
	}

	// 2. SAFETY CHECK: Only allow 'active' schedules to be synced into memory
	if sched.Status != "active" {
		// Just in case it was running, kill it.
		s.CancelJob(id)
		return fmt.Errorf("[%s]>[%s]: rejected sync for schedule %s: status is '%s'", s.prefixError, fname, id, sched.Status)
	}

	// Remove any old version of this job from memory before adding the new one
	wasRunning := s.CancelJob(id)
	if wasRunning {
		slog.WarnContext(ctx, "Removed previous version of job before syncing", slog.String("scheduleId", id))
	}

	// Add the new job to memory
	if err := s.scheduleJob(sched); err != nil {
		return fmt.Errorf("[%s]>[%s] failed to add job %s to memory engine: %w", s.prefixError, fname, id, err)
	}

	slog.InfoContext(ctx, "Synced job into memory", slog.String("scheduleId", id))
	return nil
}

func (s *scheduleEngineService) Start(ctx context.Context) error {
	const fname = "Start"

	_, err := s.engine.NewJob(
		gocron.CronJob("0 3 * * *", false),
		gocron.NewTask(s.executeImageGarbageCollection, context.Background()),
	)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to register S3 GC task", slog.String("error", err.Error()))
	}

	// Register hourly in-memory Maintenance Counter & Alert Notification Job (runs at the top of each hour)
	_, err = s.engine.NewJob(
		gocron.CronJob("0 * * * *", false),
		gocron.NewTask(s.ExecuteMaintenanceCheck, context.Background()),
	)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to register hourly maintenance check task", slog.String("error", err.Error()))
	}

	// Load active jobs from Repo on startup
	jobs, err := s.scheduleEngineRepo.GetActiveSchedules(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s] failed to load active schedules during startup:  %w", s.prefixError, fname, err)
	}

	successCount := 0
	for _, job := range jobs {
		// Capture and handle the error from scheduleJob
		if err := s.scheduleJob(&job); err != nil {
			slog.ErrorContext(ctx, "Synced job into memory", slog.String("scheduleId", job.ScheduleId), slog.String("error", err.Error()))
			continue // Skip to the next job
		}
		successCount++
	}

	slog.InfoContext(ctx, fmt.Sprintf("[STARTUP] Successfully loaded %d out of %d active schedules", successCount, len(jobs)))

	// Load active maintenance notifications into local memory cache on startup
	if s.notifRepo != nil {
		maintItems, err := s.notifRepo.GetActiveMaintenanceDetails(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to load active maintenance items into cache on startup", slog.String("error", err.Error()))
		} else {
			for _, item := range maintItems {
				s.maintenanceCache.Store(item.MaintenanceId, item)
			}
			slog.InfoContext(ctx, fmt.Sprintf("[STARTUP] Successfully cached %d active maintenance schedules in memory", len(maintItems)))
		}
	}

	// Start the gocron engine
	s.engine.Start()

	// Run an initial in-memory maintenance check on startup
	go s.ExecuteMaintenanceCheck(context.Background())

	return nil
}

func (s *scheduleEngineService) Shutdown(ctx context.Context) error {
	const fname = "Stop"
	err := s.engine.ShutdownWithContext(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	return nil
}

// Add a job to the memory engine
func (s *scheduleEngineService) scheduleJob(sched *model.Schedule) error {
	const fname = "scheduleJob"
	var jobDef gocron.JobDefinition
	if sched.ScheduleType == "recurring" && sched.CronExpression != nil {
		jobDef = gocron.CronJob(*sched.CronExpression, false)
	} else {
		jobDef = gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(sched.StartTime))
	}

	job, err := s.engine.NewJob(jobDef, gocron.NewTask(func() {}))
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	s.jobRegistry.Store(sched.ScheduleId, job.ID())
	s.engine.Update(job.ID(), jobDef, gocron.NewTask(s.executeDeviceTask, context.Background(), job.ID(), sched))

	return nil
}

// Cancel a job in memory
func (s *scheduleEngineService) CancelJob(schedId string) bool {
	val, ok := s.jobRegistry.Load(schedId)
	if ok {
		jobID := val.(uuid.UUID)
		s.engine.RemoveJob(jobID)
		s.jobRegistry.Delete(schedId)
		slog.Info("Cancelled job in memory", slog.String("scheduleId", schedId))
	}
	return ok
}

// SyncMaintenance removes the old cached entry and re-caches the latest record from Postgres
func (s *scheduleEngineService) SyncMaintenance(ctx context.Context, maintenanceId int) error {
	if s.notifRepo == nil {
		return nil
	}

	// If maintenanceId <= 0, reload the entire active cache
	if maintenanceId <= 0 {
		items, err := s.notifRepo.GetActiveMaintenanceDetails(ctx)
		if err != nil {
			return err
		}
		s.maintenanceCache.Range(func(key, _ any) bool {
			s.maintenanceCache.Delete(key)
			return true
		})
		for _, item := range items {
			s.maintenanceCache.Store(item.MaintenanceId, item)
		}
		go s.ExecuteMaintenanceCheck(context.Background())
		return nil
	}

	// Remove old cached version first
	s.maintenanceCache.Delete(maintenanceId)

	// Query latest version from DB
	item, err := s.notifRepo.GetMaintenanceDetailById(ctx, maintenanceId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.InfoContext(ctx, "Maintenance item inactive or deleted; removed from memory cache", slog.Int("maintenanceId", maintenanceId))
			return nil
		}
		return err
	}

	// Re-cache updated item and evaluate in memory
	s.maintenanceCache.Store(item.MaintenanceId, *item)
	slog.InfoContext(ctx, "Re-cached maintenance schedule in memory", slog.Int("maintenanceId", maintenanceId))

	go s.evaluateCachedMaintenanceItem(context.Background(), *item, time.Now())
	return nil
}

// UnsyncMaintenance removes a deleted maintenance schedule from memory cache
func (s *scheduleEngineService) UnsyncMaintenance(maintenanceId int) bool {
	_, existed := s.maintenanceCache.LoadAndDelete(maintenanceId)
	if existed {
		slog.Info("Removed maintenance schedule from memory cache", slog.Int("maintenanceId", maintenanceId))
	}
	return existed
}

// The actual task that fires
func (s *scheduleEngineService) executeDeviceTask(ctx context.Context, jobID uuid.UUID, sched *model.Schedule) {
	now := time.Now()

	if now.Before(sched.StartTime) {
		slog.Debug("Task triggered before StartTime, skipping execution", slog.String("scheduleId", sched.ScheduleId))
		return
	}

	slog.Info("Executing scheduled task", slog.String("scheduleId", sched.ScheduleId))

	hasValidDevice := sched.DeviceId != nil && *sched.DeviceId > 0
	hasValidGroup := sched.DeviceGroupId != nil && *sched.DeviceGroupId > 0

	// XOR Check: If both are true, or both are false, it is an invalid configuration.
	if hasValidDevice == hasValidGroup {
		slog.WarnContext(ctx, "Invalid target: Schedule must have exactly one target (Device OR Group, not both/neither)", slog.String("scheduleId", sched.ScheduleId))
		return
	}

	// Helper function to manage status transitions and audit logs
	changeStatus := func(newStatus string, reason string) {
		slog.Info(fmt.Sprintf("Changing schedule status to %s: %s", newStatus, reason), slog.String("scheduleId", sched.ScheduleId))

		// Clean up the memory engine
		s.engine.RemoveJob(jobID)
		s.jobRegistry.Delete(sched.ScheduleId)

		// Update primary database status
		s.scheduleEngineRepo.UpdateStatus(ctx, sched.ScheduleId, newStatus)
		if newStatus == "completed" {
			s.scheduleEngineRepo.UpdateLastRun(ctx, sched.ScheduleId)
		}

		oldData, err1 := model.StructToDynamicJSON(map[string]any{"status": sched.Status})
		newData, err2 := model.StructToDynamicJSON(map[string]any{
			"status":     newStatus,
			"taskAction": sched.TaskAction,
		})
		if err1 != nil || err2 != nil {
			slog.ErrorContext(ctx, "Failed to parse JSON for schedule audit log", slog.String("scheduleId", sched.ScheduleId))
		}

		// Build and save the Audit Log directly
		audit := model.AuditLog{
			EntityType: "schedule",
			EntityId:   sched.ScheduleId,
			MenuType:   "Scheduler",
			Action:     model.UpdateAction,
			ChangedBy:  0, // System execution
			OldData:    oldData,
			NewData:    newData,
		}

		if err := s.auditLogRepo.Create(ctx, []model.AuditLog{audit}); err != nil {
			slog.ErrorContext(ctx, "Failed to insert schedule audit log", slog.String("error", err.Error()))
		}
	}

	// 2. Protocol Check (Determines Cancellation)
	if sched.DeviceId != nil && *sched.DeviceId > 0 {
		protocol, err := s.deviceRepo.GetDeviceProtocol(ctx, *sched.DeviceId)
		if errors.Is(err, pgx.ErrNoRows) || protocol == nil {
			changeStatus("cancelled", "No protocol found for target device")
			return
		} else if err != nil {
			slog.ErrorContext(ctx, "Failed to get device protocol", slog.String("error", err.Error()))
			return
		}
	}

	// 3. Time / Expiration Check (Determines Completion)
	isPastEndTime := sched.EndTime != nil && now.After(*sched.EndTime)

	if sched.ScheduleType == "recurring" && isPastEndTime {
		changeStatus("completed", "Recurring schedule reached end date")
	} else if sched.ScheduleType == "one_time" {
		changeStatus("completed", "One-time job finished")
	} else {
		s.scheduleEngineRepo.UpdateLastRun(ctx, sched.ScheduleId)

		newData, err := model.StructToDynamicJSON(map[string]any{
			"taskAction": sched.TaskAction,
		})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to parse JSON for schedule audit log", slog.String("scheduleId", sched.ScheduleId))
		}
		audit := model.AuditLog{
			EntityType: "schedule",
			EntityId:   sched.ScheduleId,
			MenuType:   "Scheduler",
			Action:     model.QueryAction,
			ChangedBy:  0, // System execution
			OldData:    nil,
			NewData:    newData,
		}

		if err := s.auditLogRepo.Create(ctx, []model.AuditLog{audit}); err != nil {
			slog.ErrorContext(ctx, "Failed to insert schedule audit log", slog.String("error", err.Error()))
		}
	}

	// 4. Parse the TaskActionPayload directly from the extracted inner JSON
	var actionPayload model.TaskActionPayload
	if len(sched.TaskAction) > 0 {
		if err := json.Unmarshal(sched.TaskAction, &actionPayload); err != nil {
			slog.ErrorContext(ctx, "Failed to parse task action payload", slog.String("scheduleId", sched.ScheduleId), slog.String("error", err.Error()))
			return
		}
	}

	// 5. Resolve Target Device IDs Dynamically
	var targetDeviceIds []int
	isGroupTarget := false

	if sched.DeviceGroupId != nil && *sched.DeviceGroupId > 0 {
		isGroupTarget = true
		groupMap, err := s.deviceRepo.GetDeviceIdByDeviceGroupIds(ctx, []int{*sched.DeviceGroupId})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to fetch group devices", slog.String("error", err.Error()))
			return
		}
		targetDeviceIds = groupMap[*sched.DeviceGroupId]
	} else if sched.DeviceId != nil && *sched.DeviceId > 0 {
		targetDeviceIds = []int{*sched.DeviceId}
	}

	if len(targetDeviceIds) == 0 {
		slog.WarnContext(ctx, "No devices found for schedule, aborting execution", slog.String("scheduleId", sched.ScheduleId))
		return
	}

	// 6. Fetch full device details (Protocols & Group IDs)
	devices, err := s.deviceRepo.GetDeviceForCommandByIds(ctx, targetDeviceIds)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to fetch device info for scheduled command", slog.String("error", err.Error()))
		return
	}

	if len(devices) == 0 {
		slog.WarnContext(ctx, "Schedule targeted a group with no executable physical devices", slog.String("scheduleId", sched.ScheduleId))
		return
	}

	// 7. Bundle Commands
	finalCommands := model.BuildGatewayCommands(devices, actionPayload, isGroupTarget)

	// 8. Send all commands to the Gateway
	for _, cmd := range finalCommands {
		if err := s.gatewayClient.ExecuteCommand(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "Failed to execute scheduled command via Gateway", slog.String("error", err.Error()))
		}
	}
}

func sameCalendarDay(a, b time.Time) bool {
	ay, am, ad := a.In(time.Local).Date()
	by, bm, bd := b.In(time.Local).Date()
	return ay == by && am == bm && ad == bd
}

func calendarDaysBetween(start, current time.Time) int {
	loc := time.Local
	s := start.In(loc)
	c := current.In(loc)
	sFloor := time.Date(s.Year(), s.Month(), s.Day(), 0, 0, 0, 0, loc)
	cFloor := time.Date(c.Year(), c.Month(), c.Day(), 0, 0, 0, 0, loc)
	if cFloor.Before(sFloor) {
		return 0
	}
	return int(cFloor.Sub(sFloor).Hours() / 24)
}

// ExecuteMaintenanceCheck evaluates all cached maintenance records in local memory without querying the DB
// unless a daily counter increment or due alert actually triggers.
func (s *scheduleEngineService) ExecuteMaintenanceCheck(ctx context.Context) {
	if s.notifRepo == nil {
		return
	}

	now := time.Now()
	s.maintenanceCache.Range(func(key, value any) bool {
		item, ok := value.(model.MaintenanceNotificationDetail)
		if !ok || !item.Active {
			return true
		}
		s.evaluateCachedMaintenanceItem(ctx, item, now)
		return true
	})
}

func (s *scheduleEngineService) evaluateCachedMaintenanceItem(ctx context.Context, item model.MaintenanceNotificationDetail, now time.Time) {
	if !item.Active {
		return
	}

	cacheUpdated := false

	// 1. Local in-memory check for daily day_count increment (only writes to DB once per day when date changes)
	expectedDays := calendarDaysBetween(item.StartDate, now)
	needsDayUpdate := !now.Before(item.StartDate) && (item.LastCountedAt == nil || !sameCalendarDay(*item.LastCountedAt, now) || item.DayCount < expectedDays)
	if needsDayUpdate {
		newDayCount := item.DayCount
		if item.LastCountedAt != nil && !sameCalendarDay(*item.LastCountedAt, now) {
			newDayCount = item.DayCount + 1
		}
		if expectedDays > newDayCount {
			newDayCount = expectedDays
		}
		if err := s.notifRepo.UpdateMaintenanceDayCount(ctx, item.MaintenanceId, newDayCount); err != nil {
			slog.ErrorContext(ctx, "Failed to update maintenance day count in DB",
				slog.Int("maintenanceId", item.MaintenanceId),
				slog.String("error", err.Error()),
			)
		} else {
			nowCopy := now
			item.DayCount = newDayCount
			item.LastCountedAt = &nowCopy
			cacheUpdated = true
		}
	}

	// 2. Local in-memory check for Maintenance Due or Daily Repeat Alert
	isInitialDue := item.NotifSentCount == 0 && !now.Before(item.MaintenanceDate)
	isRepeatDue := item.NotifSentCount > 0 &&
		item.NotifSentCount <= s.maintenanceNotifRepeat &&
		!now.Before(item.MaintenanceDate.Add(time.Duration(item.NotifSentCount)*24*time.Hour)) &&
		(item.LastNotifiedAt == nil || !now.Before(item.LastNotifiedAt.Add(24*time.Hour)))

	if isInitialDue || isRepeatDue {
		sentCount, marked, err := s.notifRepo.MarkMaintenanceNotified(ctx, item.MaintenanceId, s.maintenanceNotifRepeat)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to mark maintenance item as notified",
				slog.Int("maintenanceId", item.MaintenanceId),
				slog.String("error", err.Error()),
			)
		} else if marked {
			nowCopy := now
			item.NotifSentCount = sentCount
			item.Notified = true
			item.LastNotifiedAt = &nowCopy
			cacheUpdated = true

			statusLabel := "Maintenance"
			if sentCount > 1 || now.Sub(item.MaintenanceDate) >= 24*time.Hour {
				statusLabel = "Late"
			}

			// Record system audit log for maintenance alert dispatch (ChangedBy: 0 = System)
			newData, _ := model.StructToDynamicJSON(map[string]any{
				"maintenanceId":   item.MaintenanceId,
				"deviceId":        item.DeviceId,
				"deviceGroupId":   item.DeviceGroupId,
				"dayCount":        item.DayCount,
				"operatingHours":  item.DayCount * 24,
				"maintenanceDate": item.MaintenanceDate,
				"status":          statusLabel,
				"notifSentCount":  sentCount,
				"maxRepeat":       s.maintenanceNotifRepeat,
			})
			audit := model.AuditLog{
				EntityType: "maintenance_notification",
				EntityId:   strconv.Itoa(item.MaintenanceId),
				MenuType:   "Notification Maintenance",
				Action:     model.UpdateAction,
				ChangedBy:  0, // System (userId 0)
				OldData:    nil,
				NewData:    newData,
			}
			if err := s.auditLogRepo.Create(ctx, []model.AuditLog{audit}); err != nil {
				slog.ErrorContext(ctx, "Failed to insert maintenance alert audit log", slog.String("error", err.Error()))
			}

			users, err := s.notifRepo.GetActiveUsersNotif(ctx)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to fetch active notification users for maintenance alert", slog.String("error", err.Error()))
			} else if len(users) > 0 && s.notifClient != nil {
				go s.dispatchMaintenanceNotif(context.Background(), users, item, statusLabel)
			}
		}
	}

	// Remove old and re-cache updated state in local memory
	if cacheUpdated {
		s.maintenanceCache.Delete(item.MaintenanceId)
		s.maintenanceCache.Store(item.MaintenanceId, item)
	}
}

func (s *scheduleEngineService) dispatchMaintenanceNotif(ctx context.Context, users []model.UserNotificationSend, item model.MaintenanceNotificationDetail, statusLabel string) {
	var smsUsers []model.UserNotificationSend
	var emailUsers []model.UserNotificationSend
	var lineUsers []model.UserNotificationSend

	targetTypeTH := "อุปกรณ์"
	targetTypeEN := "Device"
	targetName := item.DeviceName
	if item.DeviceGroupId != nil && *item.DeviceGroupId > 0 {
		targetTypeTH = "กลุ่มอุปกรณ์"
		targetTypeEN = "Device Group"
		targetName = item.GroupName
	}

	operatingHours := item.DayCount * 24
	reasonText := strings.TrimSpace(item.Reason)
	if reasonText == "" {
		reasonText = "ถึงกำหนดตรวจซ่อมบำรุงตามรอบเวลา / Scheduled Maintenance Due"
	}

	startStr := item.StartDate.In(time.Local).Format("02/01/2006 15:00")
	maintStr := item.MaintenanceDate.In(time.Local).Format("02/01/2006 15:00")

	statusTH := "ถึงกำหนดซ่อมบำรุง (Maintenance)"
	if statusLabel == "Late" {
		repeatIdx := max(1, item.NotifSentCount-1)
		statusTH = fmt.Sprintf("เลยกำหนดซ่อมบำรุง (Late - แจ้งเตือนซ้ำครั้งที่ %d/%d)", repeatIdx, s.maintenanceNotifRepeat)
	}

	shortMsg := fmt.Sprintf(
		"[%s] แจ้งเตือนซ่อมบำรุง: %s (%d วัน / %d ชม.) กำหนด %s - %s",
		statusLabel, targetName, item.DayCount, operatingHours, maintStr, reasonText,
	)

	emailTemplate := "🔧 แจ้งเตือนชั่วโมงการทำงานและกำหนดตรวจซ่อมบำรุง (Maintenance Alert)\n" +
		"สถานะ (Status): %s\n" +
		"%s: %s\n" +
		"ระยะเวลาทำงานสะสม: %d วัน (%d ชั่วโมง)\n" +
		"วันที่เริ่มนับ: %s\n" +
		"กำหนดตรวจซ่อมบำรุง: %s\n" +
		"รายละเอียด: %s\n\n" +
		"----------------------------------------\n\n" +
		"🔧 Operating Hours & Maintenance Due Alert\n" +
		"Status: %s\n" +
		"%s: %s\n" +
		"Accumulated Operating Time: %d Days (%d Hours)\n" +
		"Start Date: %s\n" +
		"Maintenance Due Date: %s\n" +
		"Note: %s"

	emailMsg := fmt.Sprintf(
		emailTemplate,
		statusTH, targetTypeTH, targetName, item.DayCount, operatingHours, startStr, maintStr, reasonText,
		statusLabel, targetTypeEN, targetName, item.DayCount, operatingHours, startStr, maintStr, reasonText,
	)

	for _, u := range users {
		if u.SmsActive {
			uCopy := u
			uCopy.Msg = shortMsg
			smsUsers = append(smsUsers, uCopy)
		}
		if u.EmailActive {
			uCopy := u
			uCopy.Msg = emailMsg
			emailUsers = append(emailUsers, uCopy)
		}
		if u.LineActive {
			uCopy := u
			uCopy.Msg = shortMsg
			lineUsers = append(lineUsers, uCopy)
		}
	}

	if len(smsUsers) > 0 {
		if err := s.notifClient.SendSms(ctx, smsUsers); err != nil {
			slog.Error("Failed to dispatch Maintenance SMS batch", slog.String("error", err.Error()))
		}
	}

	if len(emailUsers) > 0 {
		if err := s.notifClient.SendEmail(ctx, emailUsers); err != nil {
			slog.Error("Failed to dispatch Maintenance Email batch", slog.String("error", err.Error()))
		}
	}

	if len(lineUsers) > 0 {
		if err := s.notifClient.SendLine(ctx, lineUsers); err != nil {
			slog.Error("Failed to dispatch Maintenance Line batch", slog.String("error", err.Error()))
		}
	}
}

func (s *scheduleEngineService) executeImageGarbageCollection(ctx context.Context) {
	slog.Info("Starting daily S3 Image Garbage Collection...")

	activeUrls, err := s.widgetRepo.GetActiveImageFilenames(ctx)
	if err != nil {
		slog.Error("GC: Failed to fetch active images from DB", slog.String("error", err.Error()))
		return
	}

	// Create a fast lookup map. Ensure you extract JUST the filename if the DB stores full URLs.
	activeMap := make(map[string]bool)
	for _, url := range activeUrls {
		// e.g., Extract "1234.png" from "http://example.com/file/image/1234.png"
		// You may need to use filepath.Base(url) or strings.Split to get just the filename
		filename := path.Base(url)
		activeMap[filename] = true
	}

	s3Files, err := s.s3Client.ListAllImages(ctx)
	if err != nil {
		slog.Error("GC: Failed to list images in S3", slog.String("error", err.Error()))
		return
	}

	deletedCount := 0
	for _, s3File := range s3Files {
		if !activeMap[s3File] {
			if err := s.s3Client.DeleteImage(ctx, s3File); err != nil {
				slog.Error("GC: Failed to delete orphaned image", slog.String("file", s3File), slog.String("error", err.Error()))
				continue
			}
			slog.Info("GC: Deleted orphaned image", slog.String("file", s3File))
			deletedCount++
		}
	}

	slog.Info("S3 Garbage Collection completed", slog.Int("deleted_count", deletedCount))
}
