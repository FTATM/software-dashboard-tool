package router

import (
	"github.com/FTATM/software-dashboard-tool/internal/handler"
)

type RouterHandlers struct {
	Widget         *handler.WidgetHandler
	Canvas         *handler.CanvasHandler
	WidgetType     *handler.WidgetTypeHandler
	User           *handler.UserHandler
	Device         *handler.DeviceHandler
	Schedule       *handler.ScheduleHandler
	ScheduleEngine *handler.ScheduleEngineHandler
	Role           *handler.RoleHandler
	DeviceGateway  *handler.DeviceGatewayHandler
	LogReport      *handler.LogReportHandler
	Notification   *handler.NotificationHandler
	S3File         *handler.S3FileHandler
	// Add more as the app grows...
}
