package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/FTATM/software-dashboard-tool/internal/auth"
	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type ScheduleHandler struct {
	service              model.ScheduleService
	roleService          model.RoleService
	scheduleEngineClient model.ScheduleEngineClient
}

func NewScheduleHandler(service model.ScheduleService, roleService model.RoleService, scheduleEngineClient model.ScheduleEngineClient) *ScheduleHandler {
	return &ScheduleHandler{service: service, roleService: roleService, scheduleEngineClient: scheduleEngineClient}
}

func (h *ScheduleHandler) GetAllDetail(w http.ResponseWriter, r *http.Request) {
	var res Response
	scheduleDetails, err := h.service.GetAllDetail(r.Context())
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}
	res.Data = scheduleDetails
	respondJson(w, http.StatusOK, &res)
}

func (h *ScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var res Response
	var createScheduleReq model.CreateScheduleReq
	var err error

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Scheduler",
		ActionName: "Create",
	}

	hasAccess, err := h.roleService.Access(r.Context(), acc)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if !hasAccess {
		res.Message = "t_no_access"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if err = json.NewDecoder(r.Body).Decode(&createScheduleReq); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	err = h.service.CreateSchedule(r.Context(), &createScheduleReq, authUserId)
	if err != nil {
		if errors.Is(err, model.ErrInvalidBody) {
			res.Message = model.ErrInvalidBody.Error()
			respondJson(w, http.StatusBadRequest, &res)
		} else {
			res.Message = "Error"
			slog.ErrorContext(r.Context(), res.Message,
				slog.String("track", err.Error()),
			)
			respondJson(w, http.StatusInternalServerError, &res)
		}
		return
	}

	go func(id string) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.scheduleEngineClient.SyncSchedule(bgCtx, id); err != nil {
			slog.Error("Failed to sync schedule to engine", slog.String("scheduleId", id), slog.String("error", err.Error()))
		}
	}(createScheduleReq.ScheduleId)

	respondJson(w, http.StatusOK, &res)
}

func (h *ScheduleHandler) Update(w http.ResponseWriter, r *http.Request) {
	var res Response
	var updateScheduleReq model.UpdateScheduleReq
	var err error

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Scheduler",
		ActionName: "Update",
	}

	hasAccess, err := h.roleService.Access(r.Context(), acc)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if !hasAccess {
		res.Message = "t_no_access"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if err = json.NewDecoder(r.Body).Decode(&updateScheduleReq); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	err = h.service.UpdateSchedule(r.Context(), &updateScheduleReq, authUserId)
	if err != nil {
		if errors.Is(err, model.ErrInvalidBody) {
			res.Message = model.ErrInvalidBody.Error()
			respondJson(w, http.StatusBadRequest, &res)
		} else {
			res.Message = "Error"
			slog.ErrorContext(r.Context(), res.Message,
				slog.String("track", err.Error()),
			)
			respondJson(w, http.StatusInternalServerError, &res)
		}
		return
	}

	go func(id string) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.scheduleEngineClient.SyncSchedule(bgCtx, id); err != nil {
			slog.Error("Failed to sync schedule to engine", slog.String("scheduleId", id), slog.String("error", err.Error()))
		}
	}(updateScheduleReq.ScheduleId)

	respondJson(w, http.StatusOK, &res)
}

func (h *ScheduleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	var res Response
	var err error

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Scheduler",
		ActionName: "Delete",
	}

	hasAccess, err := h.roleService.Access(r.Context(), acc)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if !hasAccess {
		res.Message = "t_no_access"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	idStr := r.PathValue("id")
	err = h.service.DeleteSchedule(r.Context(), idStr, authUserId)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	go func(id string) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.scheduleEngineClient.UnsyncSchedule(bgCtx, id); err != nil {
			slog.Error("Failed to unsync schedule to engine", slog.String("scheduleId", id), slog.String("error", err.Error()))
		}
	}(idStr)

	respondJson(w, http.StatusOK, &res)
}
