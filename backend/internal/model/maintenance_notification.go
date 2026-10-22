package model

import "time"

type MaintenanceNotification struct {
	MaintenanceId   int        `json:"maintenanceId" db:"maintenance_id"`
	DeviceId        *int       `json:"deviceId" db:"device_id"`
	DeviceGroupId   *int       `json:"deviceGroupId" db:"device_group_id"`
	StartDate       time.Time  `json:"startDate" db:"start_date"`
	MaintenanceDate time.Time  `json:"maintenanceDate" db:"maintenance_date"`
	IntervalDays    *int       `json:"intervalDays" db:"interval_days"`
	DayCount        int        `json:"dayCount" db:"day_count"`
	Reason          string     `json:"reason" db:"reason"`
	Active          bool       `json:"active" db:"active"`
	Notified        bool       `json:"notified" db:"notified"`
	NotifSentCount  int        `json:"notifSentCount" db:"notif_sent_count"`
	LastCountedAt   *time.Time `json:"lastCountedAt" db:"last_counted_at"`
	LastNotifiedAt  *time.Time `json:"lastNotifiedAt" db:"last_notified_at"`
	CreatedAt       time.Time  `json:"-" db:"created_at"`
	UpdatedAt       time.Time  `json:"-" db:"updated_at"`
}

type MaintenanceNotificationDetail struct {
	MaintenanceNotification
	DeviceName string `json:"deviceName" db:"device_name"`
	GroupName  string `json:"groupName" db:"group_name"`
}
