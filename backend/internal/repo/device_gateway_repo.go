package repo

import (
	"context"
	"fmt"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type deviceGatewayRepo struct {
	pool        *pgxpool.Pool
	prefixError string
}

func NewDeviceGatewayRepository(pool *pgxpool.Pool) model.DeviceGatewayRepository {
	return &deviceGatewayRepo{
		pool:        pool,
		prefixError: "deviceGatewayRepo",
	}
}

func (r *deviceGatewayRepo) BulkUpsertDeviceData(ctx context.Context, data []model.DeviceData) error {
	const fname = "BulkUpsertDeviceData"

	if len(data) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s] begin tx failed: %w", r.prefixError, fname, err)
	}
	defer tx.Rollback(ctx)

	// 1. Batch update only active devices and RETURN the device_id if updated
	batch := &pgx.Batch{}
	updateQuery := `
		UPDATE device 
		SET value_data = $1, updated_value_at = now(), last_seen_at = now() 
		WHERE device_id = $2 AND active = true AND deleted_at IS NULL
		RETURNING device_id`

	for _, d := range data {
		batch.Queue(updateQuery, d.ValueData, d.DeviceId)
	}

	br := tx.SendBatch(ctx, batch)

	// Track which device IDs were successfully updated (i.e. were active)
	activeDeviceIDs := make(map[int]struct{}, len(data))
	for i := range data {
		rows, err := br.Query()
		if err != nil {
			br.Close()
			return fmt.Errorf("[%s]>[%s] batch update failed at index %d: %w", r.prefixError, fname, i, err)
		}

		var updatedID int
		if rows.Next() {
			if err := rows.Scan(&updatedID); err != nil {
				rows.Close()
				br.Close()
				return fmt.Errorf("[%s]>[%s] scan failed at index %d: %w", r.prefixError, fname, i, err)
			}
			activeDeviceIDs[updatedID] = struct{}{}
		}
		rows.Close()
	}

	if err := br.Close(); err != nil {
		return fmt.Errorf("[%s]>[%s] batch close failed: %w", r.prefixError, fname, err)
	}

	// 2. Filter data for CopyFrom — only include logs for active devices
	var rows [][]any
	for _, d := range data {
		if _, ok := activeDeviceIDs[d.DeviceId]; ok {
			rows = append(rows, []any{d.DeviceId, d.ValueData, d.ReceivedAt})
		}
	}

	// If no devices were active, skip CopyFrom and just commit
	if len(rows) > 0 {
		copyCount, err := tx.CopyFrom(
			ctx,
			pgx.Identifier{"device_data_log"},
			[]string{"device_id", "value_data", "received_at"},
			pgx.CopyFromRows(rows),
		)
		if err != nil {
			return fmt.Errorf("[%s]>[%s] copy failed: %w", r.prefixError, fname, err)
		}
		if int(copyCount) != len(rows) {
			return fmt.Errorf("[%s]>[%s] copy mismatch: expected %d rows, inserted %d", r.prefixError, fname, len(rows), copyCount)
		}
	}

	// 3. Commit
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s] commit failed: %w", r.prefixError, fname, err)
	}

	return nil
}

func (r *deviceGatewayRepo) UpdateLastSeen(ctx context.Context, deviceId int) error {
	const fname = "UpdateLastSeen"

	query := `UPDATE device SET last_seen_at = now() WHERE device_id = $1`

	result, err := r.pool.Exec(ctx, query, deviceId)

	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}

	return nil
}

func (r *deviceGatewayRepo) GetDeviceIdByName(ctx context.Context, deviceName string) (int, error) {
	const fname = "GetDeviceIdByName"
	var deviceId int
	query := `
		SELECT 
			device_id
		FROM device
		WHERE 
			device_name = $1 AND
			(active = true AND deleted_at IS NULL)
	`

	err := r.pool.QueryRow(ctx, query, deviceName).Scan(&deviceId)

	if err != nil {
		return 0, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return deviceId, nil
}

func (r *deviceGatewayRepo) GetDeviceIdByGroupName(ctx context.Context, groupName string) ([]model.DeviceGroupData, error) {
	const fname = "GetDeviceIdByGroupName"
	query := `
		SELECT 
    		dg.group_id,
    		dg.group_name,
    		COALESCE(dg.protocol::text, '') AS protocol,
    		COALESCE(array_agg(d.device_id) FILTER (WHERE d.device_id IS NOT NULL), '{}') AS device_id_s
		FROM device_group dg
		LEFT JOIN device_group_map dgm on dg.group_id = dgm.group_id
		LEFT JOIN device d on dgm.device_id = d.device_id 
    		AND d.active = true 
    		AND d.deleted_at IS NULL
			AND d.ref_device_id IS NULL
		WHERE 
    		dg.group_name = $1
		GROUP BY 
    		dg.group_id, 
    		dg.group_name, 
    		dg.protocol
    `

	rows, err := r.pool.Query(ctx, query, groupName)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	deviceGroupData, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.DeviceGroupData])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return deviceGroupData, nil
}

func (r *deviceGatewayRepo) GetDeviceNameById(ctx context.Context, deviceId int) (string, error) {
	const fname = "GetDeviceNameById"
	var deviceName string
	query := `
		SELECT 
			device_name
		FROM device
		WHERE 
			device_id = $1 AND
			(active = true AND deleted_at IS NULL)
	`

	err := r.pool.QueryRow(ctx, query, deviceId).Scan(&deviceName)

	if err != nil {
		return "", fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return deviceName, nil
}

func (r *deviceGatewayRepo) GetDeviceGroupNameById(ctx context.Context, groupId int) (string, error) {
	const fname = "GetDeviceGroupNameById"
	var deviceName string
	query := `
		SELECT 
			group_name
		FROM device_group
		WHERE 
			group_id = $1
	`

	err := r.pool.QueryRow(ctx, query, groupId).Scan(&deviceName)

	if err != nil {
		return "", fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return deviceName, nil
}

func (r *deviceGatewayRepo) GetDeviceInfoByName(ctx context.Context, deviceName string) (int, string, error) {
	const fname = "GetDeviceInfoByName"

	var deviceId int
	var protocol string

	query := `
		SELECT 
			device_id,
			protocol
		FROM device
		WHERE 
			device_name = $1 AND
			(active = true AND deleted_at IS NULL) AND
			ref_device_id IS NULL
	`

	// Scan both the ID and the protocol into their respective variables
	err := r.pool.QueryRow(ctx, query, deviceName).Scan(&deviceId, &protocol)

	if err != nil {
		return 0, "", fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return deviceId, protocol, nil
}
