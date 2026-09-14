package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type ScheduleEngineHandler struct {
	service model.ScheduleEngineService
}

func NewScheduleEngineHandler(service model.ScheduleEngineService) *ScheduleEngineHandler {
	return &ScheduleEngineHandler{service: service}
}

// Your main app calls this AFTER doing an INSERT or UPDATE in Postgres
func (h *ScheduleEngineHandler) SyncSchedule(w http.ResponseWriter, r *http.Request) {
	var schedReq model.SyncJobReq
	if err := json.NewDecoder(r.Body).Decode(&schedReq); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		slog.ErrorContext(r.Context(), "Error",
			slog.String("track", err.Error()),
		)
		return
	}

	if len(schedReq.ScheduleId) < 32 {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		slog.ErrorContext(r.Context(), "Error",
			slog.String("track", "Invalid UUID"),
		)
		return
	}

	err := h.service.SyncJob(context.Background(), schedReq.ScheduleId)
	if err != nil {
		http.Error(w, "Failed to sync job", http.StatusInternalServerError)
		slog.ErrorContext(r.Context(), "Error",
			slog.String("track", err.Error()),
		)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ScheduleEngineHandler) UnsyncSchedule(w http.ResponseWriter, r *http.Request) {
	schedID := r.PathValue("id")
	if len(schedID) < 32 {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	// Capture the boolean to check if it was actively running in memory
	wasRunning := h.service.CancelJob(schedID)

	if wasRunning {
		slog.InfoContext(r.Context(), "Successfully removed active job from memory",
			slog.String("schedId", schedID),
		)
	} else {
		slog.InfoContext(r.Context(), "Job was not found in memory (it may have already finished or was previously unsynced)",
			slog.String("schedId", schedID),
		)
	}

	w.WriteHeader(http.StatusOK)
}
