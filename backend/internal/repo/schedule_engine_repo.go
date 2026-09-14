package repo

import (
	"context"
	"fmt"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type scheduleEngineRepo struct {
	pool        DBTX
	prefixError string
}

func NewScheduleEngineRepository(pool *pgxpool.Pool) model.ScheduleEngineRepository {
	return &scheduleEngineRepo{pool: pool, prefixError: "scheduleEngineRepo"}
}

func (r *scheduleEngineRepo) db(ctx context.Context) DBTX {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

// Fetch a single schedule by ID after the main app inserts it
func (r *scheduleEngineRepo) GetById(ctx context.Context, id string) (*model.Schedule, error) {
	const fname = "GetByID"
	query := `SELECT schedule_id, device_id, task_action, schedule_type, start_time, end_time, cron_expression, device_group_id, status
	          FROM schedule WHERE schedule_id = $1`

	var s model.Schedule
	err := r.db(ctx).QueryRow(ctx, query, id).Scan(&s.ScheduleId, &s.DeviceId, &s.TaskAction, &s.ScheduleType, &s.StartTime, &s.EndTime, &s.CronExpression, &s.DeviceGroupId, &s.Status)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return &s, nil
}

func (r *scheduleEngineRepo) UpdateStatus(ctx context.Context, id, status string) error {
	const fname = "UpdateStatus"
	_, err := r.db(ctx).Exec(ctx, `UPDATE schedule SET status = $1, updated_at = now() WHERE schedule_id = $2`, status, id)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return nil
}

func (r *scheduleEngineRepo) UpdateLastRun(ctx context.Context, id string) error {
	const fname = "UpdateLastRun"
	_, err := r.db(ctx).Exec(ctx, `UPDATE schedule SET last_run_at = now() WHERE schedule_id = $1`, id)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return nil
}

func (r *scheduleEngineRepo) GetActiveSchedules(ctx context.Context) ([]model.Schedule, error) {
	const fname = "GetActiveSchedules"
	query := `SELECT schedule_id, device_id, task_action, schedule_type, start_time, end_time, cron_expression, device_group_id
	          FROM schedule WHERE status = 'active'`

	rows, err := r.db(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	defer rows.Close()

	schedules, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Schedule])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return schedules, nil
}
