package model

import (
	"context"
)

type NotificationRepository interface {

	// --- User Notif ---
	GetUserNotifAllDetail(ctx context.Context) ([]UserNotificationDetail, error)
	GetUserNotifById(ctx context.Context, userId int) (*UserNotification, error)
	UpsertUserNotif(ctx context.Context, userNotif UserNotification) error
	GetActiveUsersNotif(ctx context.Context) ([]UserNotificationSend, error)
	// --- Device Rule Notif ---
	GetDeviceRuleAllDetail(ctx context.Context) ([]DeviceRuleNotificationDetail, error)
	GetDeviceRuleById(ctx context.Context, ruleId int) (*DeviceRuleNotification, error)
	CreateDeviceRule(ctx context.Context, rule *DeviceRuleNotification) error
	UpdateDeviceRule(ctx context.Context, rule DeviceRuleNotification) error
	DeleteDeviceRule(ctx context.Context, ruleId int) error
	GetActiveRulesByDeviceId(ctx context.Context, deviceId int) ([]DeviceRuleNotification, error)
	TryAcquireDeviceAlertLock(ctx context.Context, ruleId int, cooldownMinutes int) (bool, error)
}

type NotificationService interface {
	// --- User Notif ---
	GetUserNotifAllDetail(ctx context.Context) ([]UserNotificationDetail, error)
	UpsertUserNotif(ctx context.Context, update UpdateNotification, authUserId int) error
	// --- Device Rule Notif ---
	GetDeviceRuleAllDetail(ctx context.Context) ([]DeviceRuleNotificationDetail, error)
	CreateDeviceRule(ctx context.Context, rule *DeviceRuleNotification, authUserId int) error
	UpdateDeviceRule(ctx context.Context, rule DeviceRuleNotification, authUserId int) error
	DeleteDeviceRule(ctx context.Context, ruleId int, authUserId int) error
	StartDeviceRuleAlert()
	AddDeviceDataAlert(data []DeviceData)
}

type NotificationClient interface {
	SendSms(ctx context.Context, smsUser []UserNotificationSend) error
	SendEmail(ctx context.Context, emailUsers []UserNotificationSend) error
}
