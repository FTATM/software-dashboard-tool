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
    		COALESCE(n.email_active, FALSE) AS email_active,
    		COALESCE(n.sms_active, FALSE) AS sms_active
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
			sms_active
		FROM user_notification
		WHERE user_id = $1
		`
	err := r.db(ctx).QueryRow(ctx, query, userId).Scan(&userNotif.UserId, &userNotif.EmailActive, &userNotif.SmsActive)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return &userNotif, nil
}

func (r *notificationRepo) UpsertUserNotif(ctx context.Context, userNotif model.UserNotification) error {
	const fname = "UpsertUserNotif"
	query := `
	INSERT INTO user_notification (user_id, email_active, sms_active, updated_at)
	VALUES ($1, $2, $3, now())
	ON CONFLICT (user_id) 
	DO UPDATE SET 
    	email_active = EXCLUDED.email_active,
    	sms_active = EXCLUDED.sms_active,
    	updated_at = now();
	`

	rows, err := r.db(ctx).Exec(ctx, query, userNotif.UserId, userNotif.EmailActive, userNotif.SmsActive)
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
			un.email_active, 
			un.sms_active
		FROM user_notification un
		INNER JOIN "user" u ON un.user_id = u.user_id
		WHERE un.email_active = TRUE OR un.sms_active = TRUE;
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
