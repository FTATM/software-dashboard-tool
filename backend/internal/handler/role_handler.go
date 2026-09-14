package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/FTATM/software-dashboard-tool/internal/auth"
	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type RoleHandler struct {
	service model.RoleService
}

func NewRoleHandler(service model.RoleService) *RoleHandler {
	return &RoleHandler{service: service}
}

func (h *RoleHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var res Response
	var err error
	var upsertRole model.UpsertRole

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:   authUserId,
		MenuName: "Role",
	}

	if err = json.NewDecoder(r.Body).Decode(&upsertRole); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if upsertRole.RoleId == 0 {
		acc.ActionName = "Create"
	} else {
		acc.ActionName = "Update"
	}

	hasAccess, err := h.service.Access(r.Context(), acc)
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

	err = h.service.UpsertRole(r.Context(), &upsertRole, authUserId)
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

func (h *RoleHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	var res Response
	roles, err := h.service.GetAll(r.Context())
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = roles
	respondJson(w, http.StatusOK, &res)
}

func (h *RoleHandler) GetMenuAvailable(w http.ResponseWriter, r *http.Request) {
	var res Response
	menuAvailable, err := h.service.GetMenuActionAvailable(r.Context())
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = menuAvailable
	respondJson(w, http.StatusOK, &res)
}

func (h *RoleHandler) GetDetailById(w http.ResponseWriter, r *http.Request) {
	var res Response
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}
	detail, err := h.service.GetDetailById(r.Context(), id)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = detail
	respondJson(w, http.StatusOK, &res)
}

func (h *RoleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	var res Response

	// Extract Auth User ID
	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	// Verify Permission
	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Role",
		ActionName: "Delete",
	}

	hasAccess, err := h.service.Access(r.Context(), acc)
	if err != nil {
		res.Message = "Error checking access"
		slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if !hasAccess {
		res.Message = "t_no_access"
		respondJson(w, http.StatusForbidden, &res)
		return
	}

	// Extract Role ID from the URL parameter
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	// Execute Deletion
	err = h.service.DeleteRole(r.Context(), id, authUserId)
	if err != nil {
		// Pass the exact business logic error (e.g., "Cannot delete this role...") to the client
		if errors.Is(err, model.ErrInUsed) {
			res.Message = model.ErrInUsed.Error()
			respondJson(w, http.StatusBadRequest, &res)
		} else {
			res.Message = "Error"
			slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
			respondJson(w, http.StatusInternalServerError, &res)
		}
		return
	}

	res.Message = "Role deleted successfully"
	respondJson(w, http.StatusOK, &res)
}
