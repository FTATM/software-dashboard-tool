package repo

import (
	"context"
	"fmt"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type scheduleRepo struct {
	pool        DBTX
	prefixError string
}

func NewScheduleRepository(pool *pgxpool.Pool) model.ScheduleRepository {
	return &scheduleRepo{pool: pool, prefixError: "scheduleRepo"}
}

func (r *scheduleRepo) db(ctx context.Context) DBTX {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *scheduleRepo) GetById(ctx context.Context, id string) (*model.Schedule, error) {
	const fname = "GetById"
	sched := &model.Schedule{}
	query := `
	SELECT 
    	schedule_id,
    	device_id, 
    	task_action, 
    	schedule_type, 
    	status, 
    	start_time, 
    	end_time, 
    	cron_expression,
		device_group_id
	FROM 
    	schedule
	WHERE schedule_id = $1
	`
	err := r.db(ctx).QueryRow(ctx, query, id).Scan(&sched.ScheduleId, &sched.DeviceId, &sched.TaskAction, &sched.ScheduleType, &sched.Status, &sched.StartTime, &sched.EndTime, &sched.CronExpression, &sched.DeviceGroupId)

	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return sched, nil
}

func (r *scheduleRepo) GetAll(ctx context.Context) ([]model.Schedule, error) {
	const fname = "GetAll"
	query := `
	SELECT 
    	schedule_id, 
    	device_id, 
    	task_action, 
    	schedule_type, 
    	status, 
    	start_time, 
    	end_time, 
    	cron_expression,
		device_group_id
	FROM 
    	schedule
	`
	rows, err := r.db(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	schedules, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Schedule])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return schedules, nil
}
func (r *scheduleRepo) Create(ctx context.Context, sched *model.Schedule) error {
	const fname = "Create"
	query := `
			INSERT INTO schedule (device_id, task_action, schedule_type, status, start_time, end_time, cron_expression, device_group_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7,$8)
			RETURNING schedule_id
		`
	err := r.db(ctx).QueryRow(ctx, query, &sched.DeviceId, &sched.TaskAction, &sched.ScheduleType, &sched.Status, &sched.StartTime, &sched.EndTime, &sched.CronExpression, &sched.DeviceGroupId).Scan(&sched.ScheduleId)

	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return nil

}
func (r *scheduleRepo) Update(ctx context.Context, sched *model.Schedule) error {
	const fname = "Update"
	query := `
			UPDATE schedule
			SET 
    			device_id = $2, 
    			task_action = $3, 
    			schedule_type = $4, 
    			status = $5, 
    			start_time = $6, 
    			end_time = $7, 
    			cron_expression = $8,
				device_group_id = $9
			WHERE schedule_id = $1
		`
	result, err := r.db(ctx).Exec(ctx, query, sched.ScheduleId, sched.DeviceId, sched.TaskAction, sched.ScheduleType, sched.Status, sched.StartTime, sched.EndTime, sched.CronExpression, sched.DeviceGroupId)

	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}

	return nil

}

func (r *scheduleRepo) Delete(ctx context.Context, schedId string) error {
	const fname = "Delete"

	query := `DELETE FROM schedule WHERE schedule_id = $1`

	result, err := r.db(ctx).Exec(ctx, query, schedId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}

	return nil

}
