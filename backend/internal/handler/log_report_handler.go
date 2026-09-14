package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/xuri/excelize/v2"
)

type LogReportHandler struct {
	service model.LogReportService
}

func NewLogReportHandler(service model.LogReportService) *LogReportHandler {
	return &LogReportHandler{service: service}
}

func parseCommaSeparated(input string) []string {
	var result []string
	if input == "" {
		return result
	}
	for val := range strings.SplitSeq(input, ",") {
		if trimmed := strings.TrimSpace(val); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func (h *LogReportHandler) ExportLogs(w http.ResponseWriter, r *http.Request) {
	tab := r.URL.Query().Get("tab")
	format := r.URL.Query().Get("format")

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	toTime := time.Now()
	fromTime := toTime.Add(-24 * time.Hour)

	if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
		fromTime = t
	}
	if t, err := time.Parse(time.RFC3339, toStr); err == nil {
		toTime = t
	}

	// ⚡ Extract menuTypes, fallback to entityTypes
	menuTypes := parseCommaSeparated(r.URL.Query().Get("menuTypes"))
	entityTypes := parseCommaSeparated(r.URL.Query().Get("entityTypes"))

	filter := model.LogFilter{
		From:        fromTime,
		To:          toTime,
		Keyword:     r.URL.Query().Get("keyword"),
		MenuTypes:   menuTypes,
		EntityTypes: entityTypes,
	}

	if tab == "system" {
		h.exportSystemLogs(w, r, format, filter)
	} else if tab == "device" {
		h.exportDeviceLogs(w, r, format, filter)
	} else {
		http.Error(w, "Invalid tab parameter", http.StatusBadRequest)
	}
}

func (h *LogReportHandler) SearchLogs(w http.ResponseWriter, r *http.Request) {
	var res Response
	tab := r.URL.Query().Get("tab")

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 50
	}

	offset := (page - 1) * limit

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	toTime := time.Now()
	fromTime := toTime.Add(-24 * time.Hour)

	if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
		fromTime = t
	}
	if t, err := time.Parse(time.RFC3339, toStr); err == nil {
		toTime = t
	}

	// ⚡ Extract menuTypes, fallback to entityTypes
	menuTypes := parseCommaSeparated(r.URL.Query().Get("menuTypes"))
	entityTypes := parseCommaSeparated(r.URL.Query().Get("entityTypes"))

	sortDesc := r.URL.Query().Get("sortDesc") == "true"
	sortBy := r.URL.Query().Get("sortBy")

	filter := model.LogFilter{
		From:        fromTime,
		To:          toTime,
		Keyword:     r.URL.Query().Get("keyword"),
		MenuTypes:   menuTypes,
		EntityTypes: entityTypes,
		Limit:       limit,
		Offset:      offset,
		SortBy:      sortBy,
		SortDesc:    sortDesc,
	}

	type PaginatedResponse struct {
		Logs       any `json:"logs"`
		TotalCount int `json:"totalCount"`
	}

	switch tab {
	case "system":
		logs, err := h.service.SearchSystemLogs(r.Context(), filter)
		if err != nil {
			res.Message = "Failed to search system logs"
			slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
			respondJson(w, http.StatusInternalServerError, &res)
			return
		}

		count, err := h.service.CountSystemLogs(r.Context(), filter)
		if err != nil {
			res.Message = "Failed to count system logs"
			slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
			respondJson(w, http.StatusInternalServerError, &res)
			return
		}

		if logs == nil {
			logs = []model.AuditLogReport{}
		}

		res.Data = PaginatedResponse{Logs: logs, TotalCount: count}
		res.Message = "Success"
		respondJson(w, http.StatusOK, &res)

	case "device":
		logs, err := h.service.SearchDeviceLogs(r.Context(), filter)
		if err != nil {
			res.Message = "Failed to search device logs"
			slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
			respondJson(w, http.StatusInternalServerError, &res)
			return
		}

		count, err := h.service.CountDeviceLogs(r.Context(), filter)
		if err != nil {
			res.Message = "Failed to count device logs"
			slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
			respondJson(w, http.StatusInternalServerError, &res)
			return
		}

		if logs == nil {
			logs = []model.DeviceDataLogReport{}
		}

		res.Data = PaginatedResponse{Logs: logs, TotalCount: count}
		res.Message = "Success"
		respondJson(w, http.StatusOK, &res)

	default:
		res.Message = "Invalid tab parameter"
		respondJson(w, http.StatusBadRequest, &res)
	}
}

// ⚡ Added: Endpoint for GET /logreport/getmenutypes
func (h *LogReportHandler) GetMenuTypes(w http.ResponseWriter, r *http.Request) {
	var res Response
	types, err := h.service.GetAuditLogMenuTypes(r.Context())
	if err != nil {
		res.Message = "Failed to fetch menu types"
		slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if types == nil {
		types = []string{}
	}

	res.Data = types
	res.Message = "Success"
	respondJson(w, http.StatusOK, &res)
}

func (h *LogReportHandler) GetEntityTypes(w http.ResponseWriter, r *http.Request) {
	var res Response
	types, err := h.service.GetAuditLogEntityTypes(r.Context())
	if err != nil {
		res.Message = "Failed to fetch entity types"
		slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if types == nil {
		types = []string{}
	}

	res.Data = types
	res.Message = "Success"
	respondJson(w, http.StatusOK, &res)
}

// --- SYSTEM LOGS EXPORTER ---
func (h *LogReportHandler) exportSystemLogs(w http.ResponseWriter, r *http.Request, format string, filter model.LogFilter) {
	var res Response
	logs, err := h.service.GetSystemLogsForExport(r.Context(), filter)
	if err != nil {
		res.Message = "Failed to fetch system logs"
		slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment;filename=system_logs.csv")
		writer := csv.NewWriter(w)
		// ⚡ Added Menu column
		writer.Write([]string{"Timestamp", "Action", "Menu", "Entity Type", "Entity ID", "User"})

		for _, log := range logs {
			writer.Write([]string{
				log.CreatedAt.Format(time.RFC3339),
				log.Action,
				log.MenuType,
				log.EntityType,
				log.EntityId,
				log.Username,
			})
		}
		writer.Flush()

	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment;filename=system_logs.json")

		var exportData []struct {
			Timestamp  string `json:"timestamp"`
			Action     string `json:"action"`
			MenuType   string `json:"menuType"`
			EntityType string `json:"entityType"`
			EntityId   string `json:"entityId"`
			User       string `json:"user"`
		}

		for _, log := range logs {
			exportData = append(exportData, struct {
				Timestamp  string `json:"timestamp"`
				Action     string `json:"action"`
				MenuType   string `json:"menuType"`
				EntityType string `json:"entityType"`
				EntityId   string `json:"entityId"`
				User       string `json:"user"`
			}{
				Timestamp:  log.CreatedAt.Format(time.RFC3339),
				Action:     log.Action,
				MenuType:   log.MenuType,
				EntityType: log.EntityType,
				EntityId:   log.EntityId,
				User:       log.Username,
			})
		}
		json.NewEncoder(w).Encode(exportData)

	case "excel":
		f := excelize.NewFile()
		// ⚡ Added Menu Column C
		f.SetCellValue("Sheet1", "A1", "Timestamp")
		f.SetCellValue("Sheet1", "B1", "Action")
		f.SetCellValue("Sheet1", "C1", "Menu")
		f.SetCellValue("Sheet1", "D1", "Entity Type")
		f.SetCellValue("Sheet1", "E1", "Entity ID")
		f.SetCellValue("Sheet1", "F1", "User")

		for i, log := range logs {
			rowNum := i + 2
			f.SetCellValue("Sheet1", fmt.Sprintf("A%d", rowNum), log.CreatedAt.Format(time.RFC3339))
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", rowNum), log.Action)
			f.SetCellValue("Sheet1", fmt.Sprintf("C%d", rowNum), log.MenuType)
			f.SetCellValue("Sheet1", fmt.Sprintf("D%d", rowNum), log.EntityType)
			f.SetCellValue("Sheet1", fmt.Sprintf("E%d", rowNum), log.EntityId)
			f.SetCellValue("Sheet1", fmt.Sprintf("F%d", rowNum), log.Username)
		}

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment;filename=system_logs.xlsx")
		f.Write(w)
	}
}

// --- DEVICE LOGS EXPORTER ---
func (h *LogReportHandler) exportDeviceLogs(w http.ResponseWriter, r *http.Request, format string, filter model.LogFilter) {
	var res Response
	logs, err := h.service.GetDeviceLogsForExport(r.Context(), filter)
	if err != nil {
		res.Message = "Failed to fetch device logs"
		slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment;filename=device_logs.csv")
		writer := csv.NewWriter(w)
		writer.Write([]string{"Timestamp", "Device ID", "Device Name", "Value"})

		for _, log := range logs {
			writer.Write([]string{
				log.ReceivedAt.Format(time.RFC3339),
				fmt.Sprintf("%d", log.DeviceId),
				log.DeviceName,
				fmt.Sprintf("%.2f", log.ValueData),
			})
		}
		writer.Flush()

	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment;filename=device_logs.json")

		var exportData []struct {
			Timestamp  string  `json:"timestamp"`
			DeviceId   int     `json:"deviceId"`
			DeviceName string  `json:"deviceName"`
			Value      float64 `json:"value"`
		}

		for _, log := range logs {
			exportData = append(exportData, struct {
				Timestamp  string  `json:"timestamp"`
				DeviceId   int     `json:"deviceId"`
				DeviceName string  `json:"deviceName"`
				Value      float64 `json:"value"`
			}{
				Timestamp:  log.ReceivedAt.Format(time.RFC3339),
				DeviceId:   log.DeviceId,
				DeviceName: log.DeviceName,
				Value:      log.ValueData,
			})
		}
		json.NewEncoder(w).Encode(exportData)

	case "excel":
		f := excelize.NewFile()
		f.SetCellValue("Sheet1", "A1", "Timestamp")
		f.SetCellValue("Sheet1", "B1", "Device ID")
		f.SetCellValue("Sheet1", "C1", "Device Name")
		f.SetCellValue("Sheet1", "D1", "Value")

		for i, log := range logs {
			rowNum := i + 2
			f.SetCellValue("Sheet1", fmt.Sprintf("A%d", rowNum), log.ReceivedAt.Format(time.RFC3339))
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", rowNum), log.DeviceId)
			f.SetCellValue("Sheet1", fmt.Sprintf("C%d", rowNum), log.DeviceName)
			f.SetCellValue("Sheet1", fmt.Sprintf("D%d", rowNum), log.ValueData)
		}

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment;filename=device_logs.xlsx")
		f.Write(w)
	}
}
