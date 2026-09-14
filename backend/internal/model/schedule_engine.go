package model

import "context"

type SyncJobReq struct {
	ScheduleId string
}

type ScheduleEngineRepository interface {
	GetById(ctx context.Context, id string) (*Schedule, error)
	UpdateStatus(ctx context.Context, id, status string) error
	UpdateLastRun(ctx context.Context, id string) error
	GetActiveSchedules(ctx context.Context) ([]Schedule, error)
}

type ScheduleEngineService interface {
	SyncJob(ctx context.Context, id string) error
	CancelJob(schedID string) bool
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type ScheduleEngineClient interface {
	SyncSchedule(ctx context.Context, scheduleID string) error
	UnsyncSchedule(ctx context.Context, scheduleID string) error
}
