package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/robfig/cron/v3"
)

type scheduleService struct {
	txManager    model.TransactionManager
	prefixError  string
	scheduleRepo model.ScheduleRepository
	auditLogRepo model.AuditLogRepository
}

func NewScheduleService(txManager model.TransactionManager, sr model.ScheduleRepository, alog model.AuditLogRepository) model.ScheduleService {
	return &scheduleService{txManager: txManager, prefixError: "scheduleService", scheduleRepo: sr, auditLogRepo: alog}
}

func (s *scheduleService) GetAllDetail(ctx context.Context) ([]model.ScheduleDetail, error) {
	const fname = "GetAllDetail"
	schedules, err := s.scheduleRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	scheduleDetails := make([]model.ScheduleDetail, 0, len(schedules))

	for _, s := range schedules {
		detail := model.ScheduleDetail{
			ScheduleId:     s.ScheduleId,
			DeviceId:       s.DeviceId,
			ScheduleType:   s.ScheduleType,
			TaskAction:     s.TaskAction,
			Status:         s.Status,
			StartTime:      s.StartTime,
			EndTime:        s.EndTime,
			CronExpression: s.CronExpression,
			DeviceGroupId:  s.DeviceGroupId,
		}
		scheduleDetails = append(scheduleDetails, detail)
	}

	return scheduleDetails, nil

}

func (s *scheduleService) CreateSchedule(ctx context.Context, createdSched *model.CreateScheduleReq, userId int) error {
	const fname = "CreateSchedule"
	var err error

	sched := model.Schedule{
		DeviceId:       createdSched.DeviceId,
		DeviceGroupId:  createdSched.DeviceGroupId,
		TaskAction:     createdSched.TaskAction,
		ScheduleType:   createdSched.ScheduleType,
		Status:         createdSched.Status,
		StartTime:      createdSched.StartTime,
		EndTime:        createdSched.EndTime,
		CronExpression: createdSched.CronExpression,
	}

	if sched.ScheduleType == "recurring" {
		if sched.CronExpression == nil || *sched.CronExpression == "" {
			return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, model.ErrInvalidBody)
		}

		err = validateExpression(*sched.CronExpression)
		if err != nil {
			return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	if err = s.scheduleRepo.Create(ctx, &sched); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, 1)
	newData, err := model.StructToDynamicJSON(sched)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "schedule",
		EntityId:   sched.ScheduleId,
		MenuType:   "Scheduler",
		Action:     model.CreateAction,
		ChangedBy:  userId,
		OldData:    nil,
		NewData:    newData,
	}

	auditlogs = append(auditlogs, audit)

	if err = s.auditLogRepo.Create(ctx, auditlogs); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	createdSched.ScheduleId = sched.ScheduleId // for sync sched service

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil

}

func (s *scheduleService) UpdateSchedule(ctx context.Context, updateSchedReq *model.UpdateScheduleReq, userId int) error {
	const fname = "UpdateSchedule"
	var err error

	sched := model.Schedule{
		ScheduleId:     updateSchedReq.ScheduleId,
		DeviceId:       updateSchedReq.DeviceId,
		DeviceGroupId:  updateSchedReq.DeviceGroupId,
		TaskAction:     updateSchedReq.TaskAction,
		ScheduleType:   updateSchedReq.ScheduleType,
		Status:         updateSchedReq.Status,
		StartTime:      updateSchedReq.StartTime,
		EndTime:        updateSchedReq.EndTime,
		CronExpression: updateSchedReq.CronExpression,
	}

	oldSched, err := s.scheduleRepo.GetById(ctx, updateSchedReq.ScheduleId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	// check same then return
	if oldSched.IsSame(sched) {
		return nil
	}

	if sched.ScheduleType == "recurring" {
		if sched.CronExpression == nil || *sched.CronExpression == "" {
			return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, model.ErrInvalidBody)
		}

		err = validateExpression(*sched.CronExpression)
		if err != nil {
			return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	if err = s.scheduleRepo.Update(ctx, &sched); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, 1)
	oldData, err := model.StructToDynamicJSON(oldSched)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	newData, err := model.StructToDynamicJSON(sched)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "schedule",
		EntityId:   sched.ScheduleId,
		MenuType:   "Scheduler",
		Action:     model.UpdateAction,
		ChangedBy:  userId,
		OldData:    oldData,
		NewData:    newData,
	}
	auditlogs = append(auditlogs, audit)

	if err = s.auditLogRepo.Create(ctx, auditlogs); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *scheduleService) DeleteSchedule(ctx context.Context, schedId string, authUserId int) error {
	const fname = "DeleteSchedule"
	var err error

	sched, err := s.scheduleRepo.GetById(ctx, schedId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	err = s.scheduleRepo.Delete(ctx, schedId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	oldData, err := model.StructToDynamicJSON(sched)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	audit := model.AuditLog{
		EntityType: "schedule",
		EntityId:   schedId,
		MenuType:   "Scheduler",
		Action:     model.DeleteAction,
		ChangedBy:  authUserId,
		OldData:    oldData,
		NewData:    nil,
	}

	if err = s.auditLogRepo.Create(tx.Context(), []model.AuditLog{audit}); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func validateExpression(expr string) error {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return model.ErrInvalidBody
	}

	_, err := cron.ParseStandard(expr)
	if err != nil {
		// Wrap original error so you don't lose parser feedback (e.g., invalid range, bad field count)
		return fmt.Errorf("%w: %v", model.ErrInvalidBody, err)
	}

	return nil
}
