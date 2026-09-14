package model

import (
	"context"
	"time"
)

// for app setting log
type AuditLog struct {
	Id         int         `json:"id" db:"id"`
	EntityType string      `json:"entityType" db:"entity_type"` // e.g., 'device', 'user', 'widget'
	EntityId   string      `json:"entityId" db:"entity_id"`
	MenuType   string      `json:"menuType" db:"menu_type"`
	Action     string      `json:"action" db:"action"` // 'CREATE', 'UPDATE', 'DELETE'
	ChangedBy  int         `json:"changedBy" db:"changed_by"`
	OldData    DynamicJSON `json:"-" db:"old_data"` // What it looked like before (null if INSERT)
	NewData    DynamicJSON `json:"-" db:"new_data"` // What it looks like now (null if DELETE)
	CreatedAt  time.Time   `json:"createdAt" db:"created_at"`
}

type AuditLogReport struct {
	AuditLog
	Username string `json:"username" db:"username"`
}

// action
const (
	CreateAction string = "CREATE"
	UpdateAction string = "UPDATE"
	DeleteAction string = "DELETE"
	QueryAction  string = "QUERY"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log []AuditLog) error
}
