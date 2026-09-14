package model

import (
	"context"
	"time"
)

// This model used as log report show no real entities
// Holds the query parameters from the Vue frontend
type LogFilter struct {
	From        time.Time
	To          time.Time
	Keyword     string
	MenuTypes   []string // ⚡ Added for menu_type filtering
	EntityTypes []string // Retained for backwards compatibility
	Limit       int
	Offset      int
	SortBy      string
	SortDesc    bool
}

type LogReportRepository interface {
	GetSystemLogsForExport(ctx context.Context, filter LogFilter) ([]AuditLogReport, error)
	GetDeviceLogsForExport(ctx context.Context, filter LogFilter) ([]DeviceDataLogReport, error)
	SearchSystemLogs(ctx context.Context, filter LogFilter) ([]AuditLogReport, error)
	SearchDeviceLogs(ctx context.Context, filter LogFilter) ([]DeviceDataLogReport, error)
	CountSystemLogs(ctx context.Context, filter LogFilter) (int, error)
	CountDeviceLogs(ctx context.Context, filter LogFilter) (int, error)
	GetAuditLogMenuTypes(ctx context.Context) ([]string, error) // ⚡ Added
	GetAuditLogEntityTypes(ctx context.Context) ([]string, error)
}

type LogReportService interface {
	GetSystemLogsForExport(ctx context.Context, filter LogFilter) ([]AuditLogReport, error)
	GetDeviceLogsForExport(ctx context.Context, filter LogFilter) ([]DeviceDataLogReport, error)
	SearchSystemLogs(ctx context.Context, filter LogFilter) ([]AuditLogReport, error)
	SearchDeviceLogs(ctx context.Context, filter LogFilter) ([]DeviceDataLogReport, error)
	CountSystemLogs(ctx context.Context, filter LogFilter) (int, error)
	CountDeviceLogs(ctx context.Context, filter LogFilter) (int, error)
	GetAuditLogMenuTypes(ctx context.Context) ([]string, error) // ⚡ Added
	GetAuditLogEntityTypes(ctx context.Context) ([]string, error)
}
