package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"path"
	"sync"
	"time"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type scheduleEngineService struct {
	scheduleEngineRepo model.ScheduleEngineRepository
	prefixError        string
	engine             gocron.Scheduler
	deviceRepo         model.DeviceRepository
	gatewayClient      model.DeviceGatewayClient
	jobRegistry        sync.Map
	widgetRepo         model.WidgetRepository
	s3Client           model.S3Client
	auditLogRepo       model.AuditLogRepository
}

func NewSchedulerEngineService(ser model.ScheduleEngineRepository, dr model.DeviceRepository, gc model.DeviceGatewayClient, widgetRepo model.WidgetRepository, s3Client model.S3Client, auditLogRepo model.AuditLogRepository) (model.ScheduleEngineService, error) {
	engine, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}
	return &scheduleEngineService{
		scheduleEngineRepo: ser,
		deviceRepo:         dr,
		gatewayClient:      gc,
		engine:             engine,
		prefixError:        "scheduleEngineService",
		widgetRepo:         widgetRepo,
		s3Client:           s3Client,
		auditLogRepo:       auditLogRepo,
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

	// Start the gocron engine
	s.engine.Start()

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

		// A. Clean up the memory engine
		s.engine.RemoveJob(jobID)
		s.jobRegistry.Delete(sched.ScheduleId)

		// B. Update primary database status
		s.scheduleEngineRepo.UpdateStatus(ctx, sched.ScheduleId, newStatus)
		if newStatus == "completed" {
			s.scheduleEngineRepo.UpdateLastRun(ctx, sched.ScheduleId)
		}

		// C. Safely generate JSON for the Audit Log
		oldData, err1 := model.StructToDynamicJSON(map[string]any{"status": sched.Status})
		newData, err2 := model.StructToDynamicJSON(map[string]any{"status": newStatus})
		if err1 != nil || err2 != nil {
			slog.ErrorContext(ctx, "Failed to parse JSON for schedule audit log", slog.String("scheduleId", sched.ScheduleId))
		}

		// D. Build and save the Audit Log directly
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
			// Cancel if a specifically targeted single device has no protocol
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
	}

	// 4. Parse the TaskActionPayload
	var actionPayload model.TaskActionPayload
	if len(sched.TaskAction) > 0 {
		if err := json.Unmarshal(sched.TaskAction, &actionPayload); err != nil {
			slog.ErrorContext(ctx, "Failed to parse task action", slog.String("error", err.Error()))
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

	// 7. Bundle Commands
	finalCommands := model.BuildGatewayCommands(devices, actionPayload, isGroupTarget)

	// 8. Send all commands to the Gateway
	for _, cmd := range finalCommands {
		if err := s.gatewayClient.ExecuteCommand(ctx, cmd); err != nil {
			slog.ErrorContext(ctx, "Failed to execute scheduled command via Gateway", slog.String("error", err.Error()))
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
