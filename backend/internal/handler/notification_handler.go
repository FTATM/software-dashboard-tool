package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/FTATM/software-dashboard-tool/internal/auth"
	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type NotificationHandler struct {
	service     model.NotificationService
	roleService model.RoleService
}

func NewNotificationHandler(service model.NotificationService, roleService model.RoleService) *NotificationHandler {
	return &NotificationHandler{service: service, roleService: roleService}
}

func (h *NotificationHandler) GetUserNotifAllDetail(w http.ResponseWriter, r *http.Request) {
	var res Response
	userNotif, err := h.service.GetUserNotifAllDetail(r.Context())
	if err != nil {
		res.Message = "Error"
		respondJson(w, http.StatusInternalServerError, &res)
	}

	res.Data = userNotif
	respondJson(w, http.StatusOK, &res)
}

func (h *NotificationHandler) UpsertUserNotif(w http.ResponseWriter, r *http.Request) {
	var res Response
	var updateNotif model.UpdateNotification

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Notification User",
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

	if err = json.NewDecoder(r.Body).Decode(&updateNotif); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}
	err = h.service.UpsertUserNotif(r.Context(), updateNotif, authUserId)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	respondJson(w, http.StatusOK, &res)
}

func (h *NotificationHandler) GetDeviceRuleAllDetail(w http.ResponseWriter, r *http.Request) {
	var res Response
	rules, err := h.service.GetDeviceRuleAllDetail(r.Context())
	if err != nil {
		res.Message = "Error fetching rules"
		slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if rules == nil {
		rules = []model.DeviceRuleNotificationDetail{}
	}

	res.Data = rules
	respondJson(w, http.StatusOK, &res)
}

func (h *NotificationHandler) CreateDeviceRule(w http.ResponseWriter, r *http.Request) {
	var res Response
	var rule model.DeviceRuleNotification

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Notification Device",
		ActionName: "Create",
	}

	hasAccess, err := h.roleService.Access(r.Context(), acc)
	if err != nil || !hasAccess {
		res.Message = "t_no_access"
		respondJson(w, http.StatusForbidden, &res)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if err := h.service.CreateDeviceRule(r.Context(), &rule, authUserId); err != nil {
		res.Message = "Failed to create rule"
		slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	respondJson(w, http.StatusOK, &res)
}

func (h *NotificationHandler) UpdateDeviceRule(w http.ResponseWriter, r *http.Request) {
	var res Response
	var rule model.DeviceRuleNotification

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Notification Device",
		ActionName: "Update",
	}

	hasAccess, err := h.roleService.Access(r.Context(), acc)
	if err != nil || !hasAccess {
		res.Message = "t_no_access"
		respondJson(w, http.StatusForbidden, &res)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if err := h.service.UpdateDeviceRule(r.Context(), rule, authUserId); err != nil {
		res.Message = "Failed to update rule"
		slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	respondJson(w, http.StatusOK, &res)
}

func (h *NotificationHandler) DeleteDeviceRule(w http.ResponseWriter, r *http.Request) {
	var res Response

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Notification Device",
		ActionName: "Delete",
	}

	hasAccess, err := h.roleService.Access(r.Context(), acc)
	if err != nil || !hasAccess {
		res.Message = "t_no_access"
		respondJson(w, http.StatusForbidden, &res)
		return
	}

	idStr := r.PathValue("id")
	ruleId, err := strconv.Atoi(idStr)
	if err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if err := h.service.DeleteDeviceRule(r.Context(), ruleId, authUserId); err != nil {
		res.Message = "Failed to delete rule"
		slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	respondJson(w, http.StatusOK, &res)
}
