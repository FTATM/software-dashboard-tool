package model

import (
	"context"
	"time"
)

type Schedule struct {
	ScheduleId     string      `json:"scheduleId" db:"schedule_id"`
	DeviceId       *int        `json:"deviceId" db:"device_id"`
	DeviceGroupId  *int        `json:"deviceGroupId" db:"device_group_id"`
	TaskAction     DynamicJSON `json:"taskAction" db:"task_action"`
	ScheduleType   string      `json:"scheduleType" db:"schedule_type"`
	Status         string      `json:"status" db:"status"`
	StartTime      time.Time   `json:"startTime" db:"start_time"`
	EndTime        *time.Time  `json:"endTime" db:"end_time"`
	CronExpression *string     `json:"cronExpression" db:"cron_expression"`
	LastRunAt      *time.Time  `json:"lastRunAt" db:"last_run_at"`
	CreatedAt      time.Time   `json:"createdAt" db:"created_at"`
	UpdatedAt      *time.Time  `json:"updatedAt" db:"updated_at"`
}

func (s *Schedule) IsSame(req Schedule) bool {
	if s == nil {
		return false
	}

	return s.ScheduleId == req.ScheduleId &&
		s.DeviceId == req.DeviceId &&
		s.DeviceGroupId == req.DeviceGroupId &&
		s.TaskAction.AreJSONsEqual(req.TaskAction) &&
		s.ScheduleType == req.ScheduleType &&
		s.Status == req.Status &&
		s.StartTime.Equal(req.StartTime) &&
		CompareTimePtrs(s.EndTime, req.EndTime) &&
		s.CronExpression == req.CronExpression
}

type ScheduleDetail struct {
	ScheduleId     string      `json:"scheduleId"`
	DeviceId       *int        `json:"deviceId"`
	DeviceGroupId  *int        `json:"deviceGroupId"`
	TaskAction     DynamicJSON `json:"taskAction"`
	ScheduleType   string      `json:"scheduleType"`
	Status         string      `json:"status"`
	StartTime      time.Time   `json:"startTime"`
	EndTime        *time.Time  `json:"endTime"`
	CronExpression *string     `json:"cronExpression"`
}

type CreateScheduleReq struct {
	ScheduleId     string      `json:"scheduleId"`
	DeviceId       *int        `json:"deviceId"`
	DeviceGroupId  *int        `json:"deviceGroupId"`
	TaskAction     DynamicJSON `json:"taskAction"`
	ScheduleType   string      `json:"scheduleType"`
	Status         string      `json:"status"`
	StartTime      time.Time   `json:"startTime"`
	EndTime        *time.Time  `json:"endTime"`
	CronExpression *string     `json:"cronExpression"`
}

type UpdateScheduleReq struct {
	ScheduleId     string      `json:"scheduleId"`
	DeviceId       *int        `json:"deviceId"`
	DeviceGroupId  *int        `json:"deviceGroupId"`
	TaskAction     DynamicJSON `json:"taskAction"`
	ScheduleType   string      `json:"scheduleType"`
	Status         string      `json:"status"`
	StartTime      time.Time   `json:"startTime"`
	EndTime        *time.Time  `json:"endTime"`
	CronExpression *string     `json:"cronExpression"`
}

type ScheduleRepository interface {
	GetById(ctx context.Context, id string) (*Schedule, error)
	GetAll(ctx context.Context) ([]Schedule, error)
	Create(ctx context.Context, sched *Schedule) error
	Update(ctx context.Context, sched *Schedule) error
	Delete(ctx context.Context, schedId string) error
}

type ScheduleService interface {
	GetAllDetail(ctx context.Context) ([]ScheduleDetail, error)
	CreateSchedule(ctx context.Context, sched *CreateScheduleReq, userId int) error
	UpdateSchedule(ctx context.Context, sched *UpdateScheduleReq, userId int) error
	DeleteSchedule(ctx context.Context, schedId string, userId int) error
}
