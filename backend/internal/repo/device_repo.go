package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type deviceRepo struct {
	pool        DBTX
	prefixError string
}

func NewDeviceRepository(pool *pgxpool.Pool) model.DeviceRepository {
	return &deviceRepo{pool: pool, prefixError: "deviceRepo"}
}

func (r *deviceRepo) db(ctx context.Context) DBTX {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *deviceRepo) GetById(ctx context.Context, id int) (*model.Device, error) {
	const fname = "GetById"
	device := &model.Device{}
	query := `
		SELECT 
			d.device_id, 
			d.device_name, 
			d.active, 
			d.protocol,
			d.ref_device_id,
			d.raw_min,
			d.raw_max,
			d.eu_min,
			d.eu_max,
			d.updated_value_at
		FROM device d
		WHERE d.device_id = $1 AND d.deleted_at IS NULL
	`
	err := r.db(ctx).QueryRow(ctx, query, id).Scan(
		&device.DeviceId,
		&device.DeviceName,
		&device.Active,
		&device.Protocol,
		&device.RefDeviceId,
		&device.RawMin,
		&device.RawMax,
		&device.EuMin,
		&device.EuMax,
		&device.UpdatedValueAt,
	)

	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return device, nil
}

func (r *deviceRepo) GetAll(ctx context.Context, active bool) ([]model.Device, error) {
	const fname = "GetAll"
	query := `
	SELECT 
		d.device_id, 
		d.device_name, 
		d.protocol,
		COALESCE(ref.value_data, d.value_data) AS value_data,
		d.active, 
		COALESCE(ref.last_seen_at, d.last_seen_at) AS last_seen_at,
		d.ref_device_id,
		d.raw_min,
		d.raw_max,
		d.eu_min,
		d.eu_max
	FROM device d
	LEFT JOIN device ref ON d.ref_device_id = ref.device_id AND ref.deleted_at IS NULL
	WHERE ($1 = false) OR ($1 = true AND d.deleted_at IS NULL)
	`
	rows, err := r.db(ctx).Query(ctx, query, active)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	devices, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Device])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return devices, nil
}

func (r *deviceRepo) GetAllName(ctx context.Context, active bool) ([]model.Device, error) {
	const fname = "GetAll"
	query := `
	SELECT 
		device_id, 
		device_name, 
        active
    FROM device
	WHERE ($1 = false) OR ($1 = true AND deleted_at IS NULL)
	`
	rows, err := r.db(ctx).Query(ctx, query, active)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	devices, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Device])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return devices, nil
}

func (r *deviceRepo) Create(ctx context.Context, devices []model.Device) error {
	const fname = "Create"
	if len(devices) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
    	INSERT INTO device (
        	device_name, protocol, active,
        	ref_device_id, raw_min, raw_max, eu_min, eu_max
    	) 
    	VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
    	RETURNING device_id
	`

	for _, device := range devices {
		batch.Queue(query,
			device.DeviceName, device.Protocol, device.Active,
			device.RefDeviceId, device.RawMin, device.RawMax, device.EuMin, device.EuMax,
		)
	}

	br := r.db(ctx).SendBatch(ctx, batch)
	defer br.Close()

	for i := range devices {
		err := br.QueryRow().Scan(&devices[i].DeviceId)
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, model.ErrDuplicate)
			} else {
				return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
			}
		}
	}

	return nil
}

func (r *deviceRepo) Update(ctx context.Context, device *model.Device) error {
	const fname = "Update"
	query := `
    	UPDATE device
    	SET 
        	active = $1,
        	protocol = $3,
        	device_name = $4,
        	ref_device_id = $5,
        	raw_min = $6,
        	raw_max = $7,
        	eu_min = $8,
        	eu_max = $9,
        	updated_at = now()
    	WHERE device_id = $2
	`
	result, err := r.db(ctx).Exec(ctx, query, device.Active, device.DeviceId, device.Protocol, device.DeviceName, device.RefDeviceId, device.RawMin, device.RawMax, device.EuMin, device.EuMax)

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		if pgErr.Code == "23505" {
			return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, model.ErrDuplicate)
		} else {
			return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
		}
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}

	return nil

}

func (r *deviceRepo) Delete(ctx context.Context, deviceId int) error {
	const fname = "Delete"
	query := `
			UPDATE device
			SET 
				active = false,
				deleted_at = now()
			WHERE device_id = $1 OR ref_device_id = $1
		`
	result, err := r.db(ctx).Exec(ctx, query, deviceId)

	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}

	return nil

}

func (r *deviceRepo) GetProtocolType(ctx context.Context) ([]string, error) {
	const fname = "GetAll"
	query := `
		SELECT unnest(enum_range(NULL::device_protocol_type)) AS protocol_name;
	`
	rows, err := r.db(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	protocols, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return protocols, nil
}

func (r *deviceRepo) GetByIdChartDeviceData(ctx context.Context, id int) (model.ChartDeviceData, error) {
	const fname = "GetByIdChartDeviceData"
	device := model.ChartDeviceData{}
	query := `
		SELECT 
			d.device_name, 
			COALESCE(ref.value_data, d.value_data)::float8 AS value_data, 
			COALESCE(ref.updated_value_at, d.updated_value_at) AS updated_value_at,
			d.raw_min, d.raw_max, d.eu_min, d.eu_max
		FROM device d
		LEFT JOIN device ref ON d.ref_device_id = ref.device_id
		WHERE d.device_id = $1 AND d.deleted_at IS NULL
	`
	err := r.db(ctx).QueryRow(ctx, query, id).Scan(
		&device.DeviceName,
		&device.ValueData,
		&device.UpdatedValueAt,
		&device.RawMin,
		&device.RawMax,
		&device.EuMin,
		&device.EuMax,
	)
	if err != nil {
		return device, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return device, nil
}

func (r *deviceRepo) CountData(ctx context.Context, deviceIds []int, fromTime, toTime time.Time) (int, error) {
	const fname = "CountData"
	query := `
		SELECT COUNT(*) 
		FROM device_data_log 
		WHERE device_id = ANY($1::int[]) 
		  AND received_at >= $2 
		  AND received_at <= $3;
	`
	var count int
	err := r.db(ctx).QueryRow(ctx, query, deviceIds, fromTime, toTime).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return count, err
}

// 2. Fetch Pure Raw Data (No bucketing, exactly as inserted)
func (r *deviceRepo) GetRawData(ctx context.Context, deviceIds []int, fromTime, toTime time.Time, limit int) ([]model.DeviceDataLog, error) {
	const fname = "GetRawData"
	query := `
		SELECT device_id, received_at, value_data
		FROM device_data_log
		WHERE device_id = ANY($1::int[]) AND received_at >= $2 AND received_at <= $3
		ORDER BY received_at ASC
		LIMIT $4;
	`
	rows, err := r.db(ctx).Query(ctx, query, deviceIds, fromTime, toTime, limit)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.DeviceDataLog])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return result, nil
}

// 3. Fetch Aggregated Data (Using time_bucket and last() to protect the browser)
func (r *deviceRepo) GetAggregatedData(ctx context.Context, deviceIds []int, fromTime, toTime time.Time, bucketInterval string) ([]model.DeviceDataLog, error) {
	const fname = "GetAggregatedData"
	query := `
		SELECT 
			device_id, 
			time_bucket($4::interval, received_at) AS received_at, 
			last(value_data, received_at)::INT AS value_data
		FROM device_data_log
		WHERE device_id = ANY($1::int[]) AND received_at >= $2 AND received_at <= $3
		GROUP BY device_id, time_bucket($4::interval, received_at)
		ORDER BY received_at ASC;
	`
	rows, err := r.db(ctx).Query(ctx, query, deviceIds, fromTime, toTime, bucketInterval)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	result, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.DeviceDataLog])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return result, nil
}

func (r *deviceRepo) GetDeviceGroupById(ctx context.Context, deviceGroupId int) (*model.DeviceGroup, error) {
	const fname = "GetDeviceGroupById"
	deviceGroup := model.DeviceGroup{}
	query := `
		SELECT 
			group_id,
			group_name,
			description,
			protocol	
    	FROM device_group
		WHERE group_id = $1
	`
	err := r.db(ctx).QueryRow(ctx, query, deviceGroupId).Scan(&deviceGroup.GroupId, &deviceGroup.GroupName, &deviceGroup.Description, &deviceGroup.Protocol)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return &deviceGroup, nil
}

func (r *deviceRepo) GetDeviceIdByDeviceGroupIds(ctx context.Context, deviceGroupIds []int) (map[int][]int, error) {
	const fname = "GetDeviceIdByDeviceGroupId"
	query := `
		SELECT 
			group_id,
			device_id
    	FROM device_group_map
		WHERE group_id = ANY($1::INT[])
	`
	rows, err := r.db(ctx).Query(ctx, query, deviceGroupIds)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	devicesGroupMaps, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.DeviceGroupMap])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	dmap := make(map[int][]int)
	for _, m := range devicesGroupMaps {
		dmap[m.GroupId] = append(dmap[m.GroupId], m.DeviceId)

	}

	return dmap, nil
}

func (r *deviceRepo) GetAllDeviceGroup(ctx context.Context) ([]model.DeviceGroup, error) {
	const fname = "GetAllDeviceGroup"
	query := `
		SELECT 
			group_id,
			group_name,
			description,
			protocol
    	FROM device_group
	`
	rows, err := r.db(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	devicesGroups, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.DeviceGroup])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return devicesGroups, nil
}

func (r *deviceRepo) CreateGroup(ctx context.Context, deviceGroup *model.DeviceGroup) error {
	const fname = "CreateGroup"
	query := `
			INSERT INTO device_group (group_name,description,protocol)
			VALUES ($1, $2, $3)
			RETURNING group_id
		`
	err := r.db(ctx).QueryRow(ctx, query, deviceGroup.GroupName, deviceGroup.Description, deviceGroup.Protocol).Scan(&deviceGroup.GroupId)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, model.ErrDuplicate)
			} else {
				return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
			}
		}
	}

	return nil

}

func (r *deviceRepo) UpdateGroup(ctx context.Context, deviceGroup *model.DeviceGroup) error {
	const fname = "UpdateGroup"
	query := `
			UPDATE device_group 
			SET group_name = $2,
				description = $3,
				protocol = $4
			WHERE group_id = $1
		`
	result, err := r.db(ctx).Exec(ctx, query, deviceGroup.GroupId, deviceGroup.GroupName, deviceGroup.Description, deviceGroup.Protocol)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}

	return nil

}

func (r *deviceRepo) DeleteGroup(ctx context.Context, deviceGroupId int) error {
	const fname = "DeleteGroup"
	// if len(ids) == 0 {
	// 	return nil
	// }

	query := `DELETE FROM device_group WHERE group_id = $1`

	result, err := r.db(ctx).Exec(ctx, query, deviceGroupId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}

	return nil
}

func (r *deviceRepo) CreateGroupMap(ctx context.Context, groupId int, deviceIds []int) error {
	const fname = "CreateGroupMap"
	if len(deviceIds) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
        INSERT INTO device_group_map (
            group_id,
            device_id
        ) 
        VALUES ($1, $2)`

	// 1. Queue all the queries
	for _, id := range deviceIds {
		batch.Queue(
			query,
			groupId,
			id,
		)
	}

	// 2. Send the batch in one network trip
	br := r.db(ctx).SendBatch(ctx, batch)
	defer br.Close() // Ensure the batch is closed

	// 3. Read the results and check for errors
	var insertedCount int64
	for i := range deviceIds {
		// Use Exec() for INSERT/UPDATE/DELETE without RETURNING
		commandTag, err := br.Exec()
		if err != nil {
			return fmt.Errorf("[%s]>[%s] failed at index %d (deviceId: %d): %w", r.prefixError, fname, i, deviceIds[i], err)
		}

		// Count how many rows were actually inserted
		insertedCount += commandTag.RowsAffected()
	}

	if int(insertedCount) != len(deviceIds) {
		return fmt.Errorf("[%s]>[%s] expected %d inserts, but got %d", r.prefixError, fname, len(deviceIds), insertedCount)
	}

	return nil
}

func (r *deviceRepo) DeleteGroupMap(ctx context.Context, groupId int, deviceIds []int) error {
	const fname = "DeleteGroupMap"
	if len(deviceIds) == 0 {
		return nil
	}

	query := `DELETE FROM device_group_map WHERE group_id = $1 AND device_id = ANY($2::INT[])`

	result, err := r.db(ctx).Exec(ctx, query, groupId, deviceIds)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	if result.RowsAffected() != int64(len(deviceIds)) {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}

	return nil
}

func (r *deviceRepo) GetByIds(ctx context.Context, id []int, active bool) ([]model.Device, error) {
	const fname = "GetByIds"
	query := `
	SELECT 
		d.device_id, 
		d.device_name, 
		d.protocol,
		COALESCE(ref.value_data, d.value_data) AS value_data,
		d.active,
		d.ref_device_id,
		d.raw_min,
		d.raw_max,
		d.eu_min,
		d.eu_max,
		d.updated_value_at
	FROM device d
	LEFT JOIN device ref ON d.ref_device_id = ref.device_id AND ref.deleted_at IS NULL
	WHERE (($1 = false) OR ($1 = true AND d.deleted_at IS NULL))
		AND d.device_id = ANY($2::INT[])
	`
	rows, err := r.db(ctx).Query(ctx, query, active, id)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	devices, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Device])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return devices, nil
}

func (r *deviceRepo) GetDeviceForCommandByIds(ctx context.Context, deviceIds []int) ([]model.CommandDeviceInfo, error) {
	const fname = "GetDeviceForCommandByIds"

	// ⚡ UPDATED: Joins through device_group_map to get the group protocol
	query := `
	SELECT 
		d.device_id, 
		d.device_name, 
		d.protocol,
		g.group_id,
		g.protocol AS group_protocol
	FROM device d
	LEFT JOIN device_group_map m ON d.device_id = m.device_id
	LEFT JOIN device_group g ON m.group_id = g.group_id
	WHERE d.device_id = ANY($1::INT[])
	  AND d.deleted_at IS NULL
	  AND d.active = true
	  AND d.ref_device_id IS NULL
	`

	rows, err := r.db(ctx).Query(ctx, query, deviceIds)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	// Use RowToStructByNameLax to map the SQL result to CommandDeviceInfo
	devices, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.CommandDeviceInfo])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return devices, nil
}

func (r *deviceRepo) GetDeviceProtocol(ctx context.Context, deviceId int) (*string, error) {
	const fname = "GetDeviceProtocol"
	var protocol *string
	query := `
		SELECT protocol
		FROM device
		WHERE device_id = $1 
	  		AND active = true 
	  		AND deleted_at IS NULL
	  		AND ref_device_id IS NULL
	`
	err := r.db(ctx).QueryRow(ctx, query, deviceId).Scan(&protocol)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return protocol, nil
}
