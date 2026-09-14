package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/FTATM/software-dashboard-tool/internal/auth"
	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type WidgetHandler struct {
	service     model.WidgetService
	roleService model.RoleService
}

func NewWidgetHandler(service model.WidgetService, roleService model.RoleService) *WidgetHandler {
	return &WidgetHandler{service: service, roleService: roleService}
}

func (h *WidgetHandler) GetById(w http.ResponseWriter, r *http.Request) {
	var res Response
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	widget, err := h.service.GetWidgetDetailById(r.Context(), id)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = widget
	respondJson(w, http.StatusOK, &res)
}

func (h *WidgetHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var err error
	var res Response
	var upsertReq model.UpsertWidgetReqest

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Canvas Design",
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

	if err = json.NewDecoder(r.Body).Decode(&upsertReq); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	err = h.service.UpsertWidget(r.Context(), &upsertReq, authUserId)
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
