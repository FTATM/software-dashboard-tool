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

type CanvasHandler struct {
	service     model.CanvasService
	roleService model.RoleService
}

func NewCanvasHandler(service model.CanvasService, roleService model.RoleService) *CanvasHandler {
	return &CanvasHandler{service: service, roleService: roleService}
}

func (h *CanvasHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	var res Response

	canvas, err := h.service.GetAllCanvas(r.Context())
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = canvas
	respondJson(w, http.StatusOK, &res)
}

func (h *CanvasHandler) GetAllCanvasRoleDetail(w http.ResponseWriter, r *http.Request) {
	var res Response

	canvas, err := h.service.GetAllCanvasRoleDetail(r.Context())
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if canvas == nil {
		canvas = make([]model.CanvasRoleDetail, 0)
	}
	res.Data = canvas
	respondJson(w, http.StatusOK, &res)
}

func (h *CanvasHandler) GetDetailById(w http.ResponseWriter, r *http.Request) {
	var res Response
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	canvasDetail, err := h.service.GetCanvasDetailById(r.Context(), id)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = canvasDetail
	respondJson(w, http.StatusOK, &res)
}

func (h *CanvasHandler) GetAllDetailByUser(w http.ResponseWriter, r *http.Request) {
	var err error
	var res Response
	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get user id"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	canvasDetails, err := h.service.GetAllCanvasDetailByUserRole(r.Context(), authUserId)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = canvasDetails
	respondJson(w, http.StatusOK, &res)
}

func (h *CanvasHandler) UpsertCanvasRole(w http.ResponseWriter, r *http.Request) {
	var err error
	var res Response
	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get user id"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Canvas Access",
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

	upsertCanvasRole := model.UpsertCanvasRole{}
	if err = json.NewDecoder(r.Body).Decode(&upsertCanvasRole); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	err = h.service.UpsertCanvasRole(r.Context(), &upsertCanvasRole, authUserId)
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

func (h *CanvasHandler) Create(w http.ResponseWriter, r *http.Request) {
	var err error
	var res Response
	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get user id"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Canvas",
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

	createCanvas := model.CreateCanvas{}
	if err = json.NewDecoder(r.Body).Decode(&createCanvas); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	err = h.service.CreateCanvas(r.Context(), &createCanvas, authUserId)
	if err != nil {
		if errors.Is(err, model.ErrDuplicate) {
			res.Message = model.ErrDuplicate.Error()
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

	respondJson(w, http.StatusOK, &res)
}

func (h *CanvasHandler) Update(w http.ResponseWriter, r *http.Request) {
	var err error
	var res Response
	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get user id"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Canvas",
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

	updateCanvas := model.UpdateCanvas{}
	if err = json.NewDecoder(r.Body).Decode(&updateCanvas); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	err = h.service.UpdateCanvas(r.Context(), &updateCanvas, authUserId)
	if err != nil {
		if errors.Is(err, model.ErrDuplicate) {
			res.Message = model.ErrDuplicate.Error()
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

	respondJson(w, http.StatusOK, &res)
}

func (h *CanvasHandler) Delete(w http.ResponseWriter, r *http.Request) {
	var err error
	var res Response
	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get user id"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Canvas",
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
	id, err := strconv.Atoi(idStr)
	if err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	err = h.service.DeleteCanvas(r.Context(), id, authUserId)
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

func (h *CanvasHandler) HandleRawQuery(w http.ResponseWriter, r *http.Request) {
	type QueryRequest struct {
		Query string `json:"query"`
	}
	var req QueryRequest
	var res Response

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get user id"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Dashboard",
		ActionName: "Query",
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if req.Query == "" {
		res.Message = "Query cannot be empty"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	results, err := h.service.ExecuteDynamicQuery(r.Context(), req.Query, authUserId)
	if err != nil {
		if errors.Is(err, model.ErrSecurityViolation) {
			slog.WarnContext(r.Context(), "Malicious query blocked",
				slog.Int("userId", authUserId),
				slog.String("query", req.Query),
				slog.String("reason", err.Error()),
			)
			res.Message = "Query execution failed: Invalid query"
			respondJson(w, http.StatusBadRequest, &res)
			return
		}
		slog.ErrorContext(r.Context(), "Dynamic query error",
			slog.String("track", err.Error()),
		)
		res.Message = "Query execution failed: Invalid query"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}
	res.Data = results
	respondJson(w, http.StatusOK, &res)
}
