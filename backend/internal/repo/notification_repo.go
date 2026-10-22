package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type notificationRepo struct {
	pool        DBTX
	prefixError string
}

func NewNotificationRepository(pool *pgxpool.Pool) model.NotificationRepository {
	if pool != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = pool.Exec(ctx, `
			ALTER TABLE IF EXISTS maintenance_notification ADD COLUMN IF NOT EXISTS interval_days INT;
			ALTER TABLE IF EXISTS maintenance_notification ADD COLUMN IF NOT EXISTS notif_sent_count INT NOT NULL DEFAULT 0;
		`)
	}
	return &notificationRepo{pool: pool, prefixError: "notificationRepo"}
}

func (r *notificationRepo) db(ctx context.Context) DBTX {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *notificationRepo) GetUserNotifAllDetail(ctx context.Context) ([]model.UserNotificationDetail, error) {
	const fname = "GetUserNotifAllDetail"
	query := `
		SELECT 
    		u.user_id,
    		u.first_name,
    		u.last_name,
    		u.username,
    		COALESCE(u.email, '') AS email,
    		COALESCE(u.tel, '') AS tel,
    		u.line_user_token,
    		COALESCE(n.email_active, FALSE) AS email_active,
    		COALESCE(n.sms_active, FALSE) AS sms_active,
			COALESCE(n.line_active, FALSE) AS line_active
		FROM "user" u
		LEFT JOIN user_notification n ON u.user_id = n.user_id
		WHERE u.deleted_at IS NULL
		ORDER BY u.user_id ASC
		`
	rows, err := r.db(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	userNotif, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.UserNotificationDetail])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return userNotif, nil
}

func (r *notificationRepo) GetUserNotifById(ctx context.Context, userId int) (*model.UserNotification, error) {
	const fname = "GetUserNotifById"
	userNotif := model.UserNotification{}
	query := `
		SELECT 
			user_id,
			email_active,
			sms_active,
			line_active
		FROM user_notification
		WHERE user_id = $1
		`
	err := r.db(ctx).QueryRow(ctx, query, userId).Scan(&userNotif.UserId, &userNotif.EmailActive, &userNotif.SmsActive, &userNotif.LineActive)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return &userNotif, nil
}

func (r *notificationRepo) UpsertUserNotif(ctx context.Context, userNotif model.UserNotification) error {
	const fname = "UpsertUserNotif"
	query := `
	INSERT INTO user_notification (user_id, email_active, sms_active, line_active, updated_at)
	VALUES ($1, $2, $3, $4, now())
	ON CONFLICT (user_id) 
	DO UPDATE SET 
    	email_active = EXCLUDED.email_active,
    	sms_active = EXCLUDED.sms_active,
    	line_active = EXCLUDED.line_active,
    	updated_at = now();
	`

	rows, err := r.db(ctx).Exec(ctx, query, userNotif.UserId, userNotif.EmailActive, userNotif.SmsActive, userNotif.LineActive)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	if rows.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}

	return nil

}

func (r *notificationRepo) GetDeviceRuleAllDetail(ctx context.Context) ([]model.DeviceRuleNotificationDetail, error) {
	const fname = "GetDeviceRuleAllDetail"
	query := `
		SELECT 
			r.rule_id, r.device_id, r.condition, r.threshold, r.active, r.reason,
			COALESCE(d.device_name, 'Unknown') AS device_name
		FROM device_rule_notification r
		LEFT JOIN device d ON r.device_id = d.device_id
		ORDER BY r.rule_id ASC
	`
	rows, err := r.db(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	rules, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.DeviceRuleNotificationDetail])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return rules, nil
}

func (r *notificationRepo) GetDeviceRuleById(ctx context.Context, ruleId int) (*model.DeviceRuleNotification, error) {
	const fname = "GetDeviceRuleById"
	var rule model.DeviceRuleNotification

	query := `
		SELECT rule_id, device_id, condition, threshold, active ,reason
		FROM device_rule_notification 
		WHERE rule_id = $1
	`

	err := r.db(ctx).QueryRow(ctx, query, ruleId).Scan(
		&rule.RuleId, &rule.DeviceId, &rule.Condition, &rule.Threshold, &rule.Active, &rule.Reason,
	)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return &rule, nil
}

func (r *notificationRepo) CreateDeviceRule(ctx context.Context, rule *model.DeviceRuleNotification) error {
	const fname = "CreateDeviceRule"
	query := `
		INSERT INTO device_rule_notification (device_id, condition, threshold, active, created_at, updated_at, reason)
		VALUES ($1, $2, $3, $4, now(), now(), $5)
		RETURNING rule_id
	`
	err := r.db(ctx).QueryRow(ctx, query, rule.DeviceId, rule.Condition, rule.Threshold, rule.Active, rule.Reason).Scan(&rule.RuleId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return nil
}

func (r *notificationRepo) UpdateDeviceRule(ctx context.Context, rule model.DeviceRuleNotification) error {
	const fname = "UpdateDeviceRule"
	query := `
		UPDATE device_rule_notification 
		SET device_id = $1, condition = $2, threshold = $3, active = $4, reason = $6, updated_at = now()
		WHERE rule_id = $5
	`
	rows, err := r.db(ctx).Exec(ctx, query, rule.DeviceId, rule.Condition, rule.Threshold, rule.Active, rule.RuleId, rule.Reason)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	if rows.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}
	return nil
}

func (r *notificationRepo) DeleteDeviceRule(ctx context.Context, ruleId int) error {
	const fname = "DeleteDeviceRule"
	query := `DELETE FROM device_rule_notification WHERE rule_id = $1`

	rows, err := r.db(ctx).Exec(ctx, query, ruleId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	if rows.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}
	return nil
}

func (r *notificationRepo) GetActiveRulesByDeviceId(ctx context.Context, deviceId int) ([]model.DeviceRuleNotification, error) {
	const fname = "GetActiveRulesByDeviceID"

	// ⚡ Use INNER JOIN to require both the Device AND the Rule to be active
	query := `
		SELECT 
			r.rule_id, 
			r.device_id, 
			r.condition, 
			r.threshold, 
			r.reason, 
			r.active
		FROM device_rule_notification r
		INNER JOIN device d ON r.device_id = d.device_id
		WHERE r.device_id = $1 
		  AND r.active = true 
		  AND d.active = true
	`

	rows, err := r.db(ctx).Query(ctx, query, deviceId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s] query failed: %w", r.prefixError, fname, err)
	}

	rules, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.DeviceRuleNotification])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s] collect rows failed: %w", r.prefixError, fname, err)
	}

	return rules, nil
}

func (r *notificationRepo) TryAcquireDeviceAlertLock(ctx context.Context, ruleId int, cooldownMinutes int) (bool, error) {
	const fname = "TryAcquireDeviceAlertLock"

	now := time.Now()
	cutoffTime := now.Add(-time.Duration(cooldownMinutes) * time.Minute)

	// ⚡ ATOMIC QUERY: Only update if last_triggered_at is null OR older than the cutoff
	query := `
        UPDATE device
        SET last_alert_triggered_at = $1
        WHERE device_id = $2
          AND (
              last_alert_triggered_at IS NULL 
              OR last_alert_triggered_at <= $3
          )
        RETURNING device_id;
    `

	var updatedId int
	err := r.db(ctx).QueryRow(ctx, query, now, ruleId, cutoffTime).Scan(&updatedId)

	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("[%s]>[%s] failed to acquire lock: %w", r.prefixError, fname, err)
	}

	return true, nil
}

func (r *notificationRepo) GetActiveUsersNotif(ctx context.Context) ([]model.UserNotificationSend, error) {
	const fname = "GetUsersNotif"
	query := `
		SELECT 
			un.user_id, 
			u.email, 
			u.tel,
			u.line_user_token,
			un.email_active, 
			un.sms_active,
			un.line_active
		FROM user_notification un
		INNER JOIN "user" u ON un.user_id = u.user_id
		WHERE un.email_active = TRUE OR un.sms_active = TRUE OR un.line_active = TRUE;
	`

	rows, err := r.db(ctx).Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usersNotif, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.UserNotificationSend])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s] collect rows failed: %w", r.prefixError, fname, err)
	}

	return usersNotif, nil
}

// --- Maintenance Notification Repository Methods ---

func (r *notificationRepo) GetMaintenanceAllDetail(ctx context.Context) ([]model.MaintenanceNotificationDetail, error) {
	const fname = "GetMaintenanceAllDetail"
	query := `
		SELECT 
			m.maintenance_id,
			m.device_id,
			m.device_group_id,
			m.start_date,
			m.maintenance_date,
			m.interval_days,
			m.day_count,
			COALESCE(m.reason, '') AS reason,
			m.active,
			m.notified,
			m.notif_sent_count,
			m.last_counted_at,
			m.last_notified_at,
			COALESCE(d.device_name, 'Unknown') AS device_name,
			COALESCE(g.group_name, 'Unknown') AS group_name
		FROM maintenance_notification m
		LEFT JOIN device d ON m.device_id = d.device_id
		LEFT JOIN device_group g ON m.device_group_id = g.group_id
		ORDER BY m.maintenance_id ASC
	`
	rows, err := r.db(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.MaintenanceNotificationDetail])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return items, nil
}

func (r *notificationRepo) GetActiveMaintenanceDetails(ctx context.Context) ([]model.MaintenanceNotificationDetail, error) {
	const fname = "GetActiveMaintenanceDetails"
	query := `
		SELECT 
			m.maintenance_id,
			m.device_id,
			m.device_group_id,
			m.start_date,
			m.maintenance_date,
			m.interval_days,
			m.day_count,
			COALESCE(m.reason, '') AS reason,
			m.active,
			m.notified,
			m.notif_sent_count,
			m.last_counted_at,
			m.last_notified_at,
			COALESCE(d.device_name, 'Unknown') AS device_name,
			COALESCE(g.group_name, 'Unknown') AS group_name
		FROM maintenance_notification m
		LEFT JOIN device d ON m.device_id = d.device_id
		LEFT JOIN device_group g ON m.device_group_id = g.group_id
		WHERE m.active = TRUE
		  AND (
			  (m.device_id IS NOT NULL AND d.active = TRUE AND d.deleted_at IS NULL)
			  OR (m.device_group_id IS NOT NULL AND g.group_id IS NOT NULL)
		  )
		ORDER BY m.maintenance_id ASC
	`
	rows, err := r.db(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.MaintenanceNotificationDetail])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return items, nil
}

func (r *notificationRepo) GetMaintenanceDetailById(ctx context.Context, maintenanceId int) (*model.MaintenanceNotificationDetail, error) {
	const fname = "GetMaintenanceDetailById"
	query := `
		SELECT 
			m.maintenance_id,
			m.device_id,
			m.device_group_id,
			m.start_date,
			m.maintenance_date,
			m.interval_days,
			m.day_count,
			COALESCE(m.reason, '') AS reason,
			m.active,
			m.notified,
			m.notif_sent_count,
			m.last_counted_at,
			m.last_notified_at,
			COALESCE(d.device_name, 'Unknown') AS device_name,
			COALESCE(g.group_name, 'Unknown') AS group_name
		FROM maintenance_notification m
		LEFT JOIN device d ON m.device_id = d.device_id
		LEFT JOIN device_group g ON m.device_group_id = g.group_id
		WHERE m.maintenance_id = $1
		  AND m.active = TRUE
		  AND (
			  (m.device_id IS NOT NULL AND d.active = TRUE AND d.deleted_at IS NULL)
			  OR (m.device_group_id IS NOT NULL AND g.group_id IS NOT NULL)
		  )
	`
	rows, err := r.db(ctx).Query(ctx, query, maintenanceId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[model.MaintenanceNotificationDetail])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return &item, nil
}

func (r *notificationRepo) GetMaintenanceById(ctx context.Context, maintenanceId int) (*model.MaintenanceNotification, error) {
	const fname = "GetMaintenanceById"
	var item model.MaintenanceNotification

	query := `
		SELECT 
			maintenance_id, device_id, device_group_id, start_date, maintenance_date, interval_days,
			day_count, COALESCE(reason, '') AS reason, active, notified, notif_sent_count, last_counted_at, last_notified_at
		FROM maintenance_notification
		WHERE maintenance_id = $1
	`

	err := r.db(ctx).QueryRow(ctx, query, maintenanceId).Scan(
		&item.MaintenanceId,
		&item.DeviceId,
		&item.DeviceGroupId,
		&item.StartDate,
		&item.MaintenanceDate,
		&item.IntervalDays,
		&item.DayCount,
		&item.Reason,
		&item.Active,
		&item.Notified,
		&item.NotifSentCount,
		&item.LastCountedAt,
		&item.LastNotifiedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return &item, nil
}

func (r *notificationRepo) CreateMaintenance(ctx context.Context, item *model.MaintenanceNotification) error {
	const fname = "CreateMaintenance"
	query := `
		INSERT INTO maintenance_notification (
			device_id, device_group_id, start_date, maintenance_date, interval_days,
			day_count, reason, active, notified, notif_sent_count, last_counted_at, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5,
			GREATEST(0, (CURRENT_DATE - ($3::timestamptz)::date)),
			$6, $7, FALSE, 0, now(), now(), now()
		)
		RETURNING maintenance_id, day_count
	`
	err := r.db(ctx).QueryRow(
		ctx,
		query,
		item.DeviceId,
		item.DeviceGroupId,
		item.StartDate,
		item.MaintenanceDate,
		item.IntervalDays,
		item.Reason,
		item.Active,
	).Scan(&item.MaintenanceId, &item.DayCount)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return nil
}

func (r *notificationRepo) UpdateMaintenance(ctx context.Context, item model.MaintenanceNotification) error {
	const fname = "UpdateMaintenance"
	query := `
		UPDATE maintenance_notification
		SET 
			device_id = $1,
			device_group_id = $2,
			start_date = $3,
			maintenance_date = $4,
			interval_days = $5,
			day_count = GREATEST(0, (CURRENT_DATE - ($3::timestamptz)::date)),
			reason = $6,
			active = $7,
			notified = CASE 
				WHEN maintenance_date <> $4::timestamptz OR start_date <> $3::timestamptz THEN FALSE
				WHEN $4::timestamptz > now() THEN FALSE
				ELSE notified
			END,
			notif_sent_count = CASE 
				WHEN maintenance_date <> $4::timestamptz OR start_date <> $3::timestamptz THEN 0
				WHEN $4::timestamptz > now() THEN 0
				ELSE notif_sent_count
			END,
			last_notified_at = CASE 
				WHEN maintenance_date <> $4::timestamptz OR start_date <> $3::timestamptz THEN NULL
				WHEN $4::timestamptz > now() THEN NULL
				ELSE last_notified_at
			END,
			last_counted_at = now(),
			updated_at = now()
		WHERE maintenance_id = $8
	`
	rows, err := r.db(ctx).Exec(
		ctx,
		query,
		item.DeviceId,
		item.DeviceGroupId,
		item.StartDate,
		item.MaintenanceDate,
		item.IntervalDays,
		item.Reason,
		item.Active,
		item.MaintenanceId,
	)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	if rows.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}
	return nil
}

func (r *notificationRepo) DeleteMaintenance(ctx context.Context, maintenanceId int) error {
	const fname = "DeleteMaintenance"
	query := `DELETE FROM maintenance_notification WHERE maintenance_id = $1`

	rows, err := r.db(ctx).Exec(ctx, query, maintenanceId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	if rows.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}
	return nil
}

func (r *notificationRepo) UpdateMaintenanceDayCount(ctx context.Context, maintenanceId int, dayCount int) error {
	const fname = "UpdateMaintenanceDayCount"
	query := `
		UPDATE maintenance_notification
		SET day_count = $1, last_counted_at = now(), updated_at = now()
		WHERE maintenance_id = $2
	`
	_, err := r.db(ctx).Exec(ctx, query, dayCount, maintenanceId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return nil
}

func (r *notificationRepo) MarkMaintenanceNotified(ctx context.Context, maintenanceId int, maxRepeat int) (int, bool, error) {
	const fname = "MarkMaintenanceNotified"
	query := `
		UPDATE maintenance_notification m
		SET 
			notified = TRUE,
			notif_sent_count = m.notif_sent_count + 1,
			last_notified_at = now(),
			updated_at = now()
		WHERE m.maintenance_id = $1
		  AND m.active = TRUE
		  AND (
			  (m.notif_sent_count = 0 AND now() >= m.maintenance_date)
			  OR (
				  m.notif_sent_count > 0
				  AND m.notif_sent_count <= $2
				  AND now() >= m.maintenance_date + (m.notif_sent_count * INTERVAL '1 day')
				  AND (m.last_notified_at IS NULL OR now() >= m.last_notified_at + INTERVAL '1 day')
			  )
		  )
		  AND (
			  (m.device_id IS NOT NULL AND EXISTS (SELECT 1 FROM device d WHERE d.device_id = m.device_id AND d.active = TRUE AND d.deleted_at IS NULL))
			  OR (m.device_group_id IS NOT NULL AND EXISTS (SELECT 1 FROM device_group g WHERE g.group_id = m.device_group_id))
		  )
		RETURNING notif_sent_count;
	`
	var newCount int
	err := r.db(ctx).QueryRow(ctx, query, maintenanceId, maxRepeat).Scan(&newCount)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("[%s]>[%s] failed to mark notified: %w", r.prefixError, fname, err)
	}
	return newCount, true, nil
}
