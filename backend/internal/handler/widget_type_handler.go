package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type WidgetTypeHandler struct {
	service model.WidgetTypeService
}

func NewWidgetTypeHandler(service model.WidgetTypeService) *WidgetTypeHandler {
	return &WidgetTypeHandler{service: service}
}

func (h *WidgetTypeHandler) GetById(w http.ResponseWriter, r *http.Request) {
	var res Response
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	widgetType, err := h.service.GetWidgetTypeById(r.Context(), id)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = widgetType
	respondJson(w, http.StatusOK, &res)
}

func (h *WidgetTypeHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	var res Response
	widgetTypes, err := h.service.GetWidgetTypeAll(r.Context())
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = widgetTypes
	respondJson(w, http.StatusOK, &res)
}
