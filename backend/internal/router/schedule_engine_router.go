package router

import (
	"net/http"
	"os"

	"github.com/FTATM/software-dashboard-tool/internal/middleware"
)

func SetupSchedule(handlers RouterHandlers) *http.ServeMux {
	mux := http.NewServeMux()

	scheduleEngine := http.NewServeMux()
	internalSecret := os.Getenv("INTERNAL_API_SECRET")

	secureMiddleware := middleware.InternalOnly(internalSecret)
	mux.Handle("/scheduleengine/", http.StripPrefix("/scheduleengine", scheduleEngine))

	syncSchedule := http.HandlerFunc(handlers.ScheduleEngine.SyncSchedule)
	unsyncSchedule := http.HandlerFunc(handlers.ScheduleEngine.UnsyncSchedule)
	syncMaintenance := http.HandlerFunc(handlers.ScheduleEngine.SyncMaintenance)
	unsyncMaintenance := http.HandlerFunc(handlers.ScheduleEngine.UnsyncMaintenance)

	scheduleEngine.Handle("POST /internal/sync", secureMiddleware(syncSchedule))
	scheduleEngine.Handle("DELETE /internal/unsync/{id}", secureMiddleware(unsyncSchedule))
	scheduleEngine.Handle("POST /internal/maintenance/sync", secureMiddleware(syncMaintenance))
	scheduleEngine.Handle("DELETE /internal/maintenance/unsync/{id}", secureMiddleware(unsyncMaintenance))
	return mux
}
