package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/jackc/pgx/v5"
)

type notificationService struct {
	txManager         model.TransactionManager
	notifRepo         model.NotificationRepository
	auditLogRepo      model.AuditLogRepository
	prefixError       string
	deviceDataAlert   chan []model.DeviceData
	notifClient       model.NotificationClient
	cooldownNotifSend int
}

func NewNotificationService(txManager model.TransactionManager, repo model.NotificationRepository, auditLogRepo model.AuditLogRepository, notifClient model.NotificationClient, dataStream chan []model.DeviceData, cooldownNotifSend int) model.NotificationService {
	return &notificationService{
		txManager:         txManager,
		notifRepo:         repo,
		auditLogRepo:      auditLogRepo,
		prefixError:       "notificationService",
		notifClient:       notifClient,
		deviceDataAlert:   dataStream,
		cooldownNotifSend: cooldownNotifSend,
	}
}

func (s *notificationService) GetUserNotifAllDetail(ctx context.Context) ([]model.UserNotificationDetail, error) {
	const fname = "GetAllDetail"
	userNotif, err := s.notifRepo.GetUserNotifAllDetail(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	return userNotif, nil
}

func (s *notificationService) UpsertUserNotif(ctx context.Context, update model.UpdateNotification, authUserId int) error {
	const fname = "UpdateUserNotification"

	userNotif := model.UserNotification{
		UserId:      update.UserId,
		SmsActive:   update.SmsActive,
		EmailActive: update.EmailActive,
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	defer tx.Rollback(ctx)

	err = s.notifRepo.UpsertUserNotif(tx.Context(), userNotif)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	newData, err := model.StructToDynamicJSON(userNotif)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "user_notification",
		EntityId:   strconv.Itoa(userNotif.UserId),
		MenuType:   "Notification User",
		Action:     model.UpdateAction,
		ChangedBy:  authUserId,
		OldData:    nil,
		NewData:    newData,
	}

	if err = s.auditLogRepo.Create(tx.Context(), []model.AuditLog{audit}); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *notificationService) GetDeviceRuleAllDetail(ctx context.Context) ([]model.DeviceRuleNotificationDetail, error) {
	const fname = "GetDeviceRuleAllDetail"
	rules, err := s.notifRepo.GetDeviceRuleAllDetail(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	return rules, nil
}

func (s *notificationService) CreateDeviceRule(ctx context.Context, rule *model.DeviceRuleNotification, authUserId int) error {
	const fname = "CreateDeviceRule"

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	defer tx.Rollback(ctx)

	if err = s.notifRepo.CreateDeviceRule(tx.Context(), rule); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	newData, _ := model.StructToDynamicJSON(rule)
	audit := model.AuditLog{
		EntityType: "device_rule_notification",
		EntityId:   strconv.Itoa(rule.RuleId),
		MenuType:   "Notification Device",
		Action:     model.CreateAction,
		ChangedBy:  authUserId,
		OldData:    nil,
		NewData:    newData,
	}

	if err = s.auditLogRepo.Create(tx.Context(), []model.AuditLog{audit}); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *notificationService) UpdateDeviceRule(ctx context.Context, rule model.DeviceRuleNotification, authUserId int) error {
	const fname = "UpdateDeviceRule"

	oldRule, err := s.notifRepo.GetDeviceRuleById(ctx, rule.RuleId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	defer tx.Rollback(ctx)

	if err = s.notifRepo.UpdateDeviceRule(tx.Context(), rule); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	oldData, _ := model.StructToDynamicJSON(oldRule)
	newData, _ := model.StructToDynamicJSON(rule)

	audit := model.AuditLog{
		EntityType: "device_rule_notification",
		EntityId:   strconv.Itoa(rule.RuleId),
		MenuType:   "Notification Device",
		Action:     model.UpdateAction,
		ChangedBy:  authUserId,
		OldData:    oldData,
		NewData:    newData,
	}

	if err = s.auditLogRepo.Create(tx.Context(), []model.AuditLog{audit}); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *notificationService) DeleteDeviceRule(ctx context.Context, ruleId int, authUserId int) error {
	const fname = "DeleteDeviceRule"

	oldRule, err := s.notifRepo.GetDeviceRuleById(ctx, ruleId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	defer tx.Rollback(ctx)

	if err = s.notifRepo.DeleteDeviceRule(tx.Context(), ruleId); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	oldData, _ := model.StructToDynamicJSON(oldRule)
	audit := model.AuditLog{
		EntityType: "device_rule_notification",
		EntityId:   strconv.Itoa(ruleId),
		MenuType:   "Notification Device",
		Action:     model.DeleteAction,
		ChangedBy:  authUserId,
		OldData:    oldData,
		NewData:    nil,
	}

	if err = s.auditLogRepo.Create(tx.Context(), []model.AuditLog{audit}); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *notificationService) StartDeviceRuleAlert() {
	for dataBatch := range s.deviceDataAlert {
		// Process each batch asynchronously
		go s.evaluateBatch(dataBatch)
	}
}

func (s *notificationService) AddDeviceDataAlert(data []model.DeviceData) {
	select {
	case s.deviceDataAlert <- data:
		// Success! The channel had room and accepted the data.
	default:
		// The channel is completely full (Database is likely lagging).
		slog.Warn("Gateway batcher notification channel is full! Dropping to prevent network block")
	}
}

func (s *notificationService) evaluateBatch(dataBatch []model.DeviceData) {
	const fname = "evaluateBatch"
	ctx := context.Background()

	for _, data := range dataBatch {
		rules, err := s.notifRepo.GetActiveRulesByDeviceId(ctx, data.DeviceId)
		if err != nil {
			slog.Debug(err.Error())
			continue
		}

		// ⚡ 1. Calculate the real decimal value immediately!
		realValue := float64(data.ValueData) / float64(model.DeviceScale)

		var triggeredRules []model.DeviceRuleNotification

		// 2. Check ALL rules using the real, unscaled decimal value
		for _, rule := range rules {
			// Cast rule.Threshold to float64 to ensure accurate decimal comparison
			if s.checkCondition(realValue, rule.Condition, float64(rule.Threshold)) {
				triggeredRules = append(triggeredRules, rule)
			}
		}

		// 3. If multiple rules triggered, pick the MOST critical one
		if len(triggeredRules) > 0 {
			mostCriticalRule := s.findMostCriticalRule(triggeredRules)

			cooldownToUse := s.cooldownNotifSend

			// Lock the device using the determined cooldown
			lockAcquired, err := s.notifRepo.TryAcquireDeviceAlertLock(
				ctx,
				data.DeviceId,
				cooldownToUse,
			)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					slog.DebugContext(ctx, fmt.Sprintf("[%s]>[%s]: No row", s.prefixError, fname))
				} else {
					slog.ErrorContext(ctx, fmt.Sprintf("[%s]>[%s]: Error lock", s.prefixError, fname))
				}
			}

			if lockAcquired {
				users, _ := s.notifRepo.GetActiveUsersNotif(ctx)
				go s.dispatchGroupedNotif(context.Background(), users, mostCriticalRule, data)
			}
		}
	}
}

// Simple logic evaluator
func (s *notificationService) checkCondition(currentValue float64, operator string, threshold float64) bool {
	switch operator {
	case ">":
		return currentValue > threshold
	case ">=":
		return currentValue >= threshold
	case "<":
		return currentValue < threshold
	case "<=":
		return currentValue <= threshold
	case "==":
		return currentValue == threshold
	case "!=":
		return currentValue != threshold
	default:
		return false
	}
}

func (s *notificationService) dispatchGroupedNotif(ctx context.Context, users []model.UserNotificationSend, rule model.DeviceRuleNotification, data model.DeviceData) {
	var smsUsers []model.UserNotificationSend
	var emailUsers []model.UserNotificationSend

	// 1. แปลงค่ากลับไปเป็นทศนิยมตามจริง (หารด้วย 1000)
	realValue := float64(data.ValueData) / float64(model.DeviceScale)

	// 2. Loop through users and group them by their active flags
	for _, u := range users {

		// 1. SMS: Ultra-compact hybrid format to avoid multi-part SMS billing
		if u.SmsActive {
			// Example output: "Alert/เตือน: Voltage-A=220.50. High Volt"
			u.Msg = fmt.Sprintf("Alert/เตือน: %s=%.2f. %s", data.DeviceName, realValue, rule.Reason)
			smsUsers = append(smsUsers, u)
		}

		// 2. Email: Full bilingual support with clear formatting
		if u.EmailActive {
			emailTemplate := "⚠️ แจ้งเตือนความผิดปกติ (System Alert)\n" +
				"อุปกรณ์ (Device): %s\n" +
				"ค่าปัจจุบัน (Value): %.2f\n" +
				"สาเหตุ (Reason): %s\n\n" +
				"----------------------------------------\n\n" +
				"⚠️ System Alert\n" +
				"Device: %s\n" +
				"Value: %.2f\n" +
				"Reason: %s"

			u.Msg = fmt.Sprintf(emailTemplate,
				data.DeviceName, realValue, rule.Reason, // Thai
				data.DeviceName, realValue, rule.Reason, // English
			)
			emailUsers = append(emailUsers, u)
		}
	}

	// 3. Dispatch to the specific clients
	if len(smsUsers) > 0 {
		if err := s.notifClient.SendSms(ctx, smsUsers); err != nil {
			slog.Error("Failed to dispatch SMS batch", slog.String("error", err.Error()))
		}
	}

	if len(emailUsers) > 0 {
		if err := s.notifClient.SendEmail(ctx, emailUsers); err != nil {
			slog.Error("Failed to dispatch Email batch", slog.String("error", err.Error()))
		}
	}
}

func (s *notificationService) findMostCriticalRule(rules []model.DeviceRuleNotification) model.DeviceRuleNotification {
	// If only one rule triggered, it is automatically the most critical
	if len(rules) == 1 {
		return rules[0]
	}

	mostCritical := rules[0]

	for _, rule := range rules {

		switch rule.Condition {
		case ">", ">=":
			// For 'greater than', a HIGHER threshold is more critical
			// (e.g., Temp > 50 is more critical than Temp > 40)
			if rule.Threshold > mostCritical.Threshold {
				mostCritical = rule
			}
		case "<", "<=":
			// For 'less than', a LOWER threshold is more critical
			// (e.g., Battery < 10 is more critical than Battery < 20)
			if rule.Threshold < mostCritical.Threshold {
				mostCritical = rule
			}
		case "==":
			// If it's an exact match, the threshold itself is the critical point.
			// It generally shouldn't conflict, but we keep the first one found.
			continue
		}
	}

	return mostCritical
}
