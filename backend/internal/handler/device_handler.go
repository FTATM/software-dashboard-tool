package handler

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/FTATM/software-dashboard-tool/internal/auth"
	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type DeviceHandler struct {
	service             model.DeviceService
	roleService         model.RoleService
	deviceGatewayClient model.DeviceGatewayClient
	auditLogRepo        model.AuditLogRepository
}

type ImportRow struct {
	DeviceName string `json:"deviceName"`
	Protocol   string `json:"protocol"`
	Active     bool   `json:"active"`
	IsValid    bool   `json:"isValid"`
	Message    string `json:"message"`
}

func NewDeviceHandler(service model.DeviceService, rs model.RoleService, deviceGatewayClient model.DeviceGatewayClient, auditLogRepo model.AuditLogRepository) *DeviceHandler {
	return &DeviceHandler{service: service, roleService: rs, deviceGatewayClient: deviceGatewayClient, auditLogRepo: auditLogRepo}
}

func (h *DeviceHandler) GetAllDetail(w http.ResponseWriter, r *http.Request) {
	var res Response
	deviceDetails, err := h.service.GetAllDeviceDetail(r.Context())
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = deviceDetails
	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var res Response
	var err error
	var devices []model.DeviceCreate

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Device",
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

	if err = json.NewDecoder(r.Body).Decode(&devices); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	err = h.service.CreateDevice(r.Context(), devices, authUserId)
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

func (h *DeviceHandler) Update(w http.ResponseWriter, r *http.Request) {
	var res Response
	var err error
	var device model.DeviceUpdate

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Device",
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

	if err = json.NewDecoder(r.Body).Decode(&device); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	userId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	err = h.service.UpdateDevice(r.Context(), &device, userId)

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

	go func(oldName string) {
		// Use a fresh context since the request context will be cancelled immediately
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.deviceGatewayClient.InvalidateDeviceCache(bgCtx, oldName); err != nil {
			slog.Error("Failed to invalidate gateway cache", slog.String("error", err.Error()))
		}
	}(device.OldName)

	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
		MenuName:   "Device",
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
	deleteDeviceId, err := strconv.Atoi(idStr)
	if err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	device, err := h.service.DeleteDevice(r.Context(), deleteDeviceId, authUserId)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if device != nil {
		go func(oldName string) {
			// Use a fresh context since the request context will be cancelled immediately
			bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := h.deviceGatewayClient.InvalidateDeviceCache(bgCtx, oldName); err != nil {
				slog.Error("Failed to invalidate gateway cache", slog.String("error", err.Error()))
			}
		}(device.DeviceName)
	}

	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceHandler) PingDevice(w http.ResponseWriter, r *http.Request) {
	var res Response

	// 1. Verify User Auth & Access
	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Device",
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

	deviceIdStr := r.URL.Query().Get("deviceId")
	if deviceIdStr == "" {
		res.Message = "Missing deviceId"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	deviceId, err := strconv.Atoi(deviceIdStr)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	connection, err := h.deviceGatewayClient.GetDeviceStatus(r.Context(), deviceId)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to Get Device Status via Gateway", slog.String("track", err.Error()))
		res.Message = "Failed to communicate with device"
		respondJson(w, http.StatusServiceUnavailable, &res)
		return
	}

	res.Data = map[string]any{
		"connection": connection,
	}
	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceHandler) TriggerCommand(w http.ResponseWriter, r *http.Request) {
	var res Response
	var req model.TriggerCommandReq

	// 1. Verify User Auth & Access
	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Unauthorized"
		respondJson(w, http.StatusUnauthorized, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Device",
		ActionName: "Command",
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

	// 2. Decode frontend payload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	// 3. Parse the DynamicJSON into the typed payload struct
	var actionPayload model.TaskActionPayload
	if len(req.TaskAction) > 0 {
		if err := json.Unmarshal(req.TaskAction, &actionPayload); err != nil {
			res.Message = "Invalid task action format"
			respondJson(w, http.StatusBadRequest, &res)
			return
		}
	}

	// 4. Fetch the target devices
	// NOTE: Make sure your DB query for this returns the Device's GroupId and the Group's Protocol!
	devices, err := h.service.GetDeviceForCommandByIds(r.Context(), req.DeviceIds)
	if err != nil {
		res.Message = "Error fetching devices"
		slog.ErrorContext(r.Context(), res.Message, slog.String("error", err.Error()))
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if len(devices) == 0 {
		res.Message = "No physical devices found to receive commands (virtual sensors cannot accept hardware commands)"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	// 5. Bundle devices by GroupId (for Gateway) and isolate single devices
	finalCommands := model.BuildGatewayCommands(devices, actionPayload, req.IsGroup)

	// 6. Send all commands to the Gateway via the client
	for _, cmd := range finalCommands {
		if err := h.deviceGatewayClient.ExecuteCommand(r.Context(), cmd); err != nil {
			slog.ErrorContext(r.Context(), "Failed to execute command via Gateway", slog.String("error", err.Error()))
			// Depending on your preference, you can return immediately or continue looping to execute the remaining devices
			res.Message = "Failed to communicate with one or more devices"
			respondJson(w, http.StatusServiceUnavailable, &res)
			return
		}
	}

	newData, err := model.StructToDynamicJSON(finalCommands)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to new data audit log trigger command", slog.String("error", err.Error()))
	}

	audit := model.AuditLog{
		EntityType: "device",
		EntityId:   strconv.Itoa(0),
		MenuType:   "Device Action",
		Action:     model.QueryAction,
		ChangedBy:  authUserId,
		OldData:    nil,
		NewData:    newData,
	}

	if err = h.auditLogRepo.Create(r.Context(), []model.AuditLog{audit}); err != nil {
		slog.ErrorContext(r.Context(), "Failed to create audit log trigger command", slog.String("error", err.Error()))
	}

	res.Message = "Commands executed successfully"
	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceHandler) GetProtocolType(w http.ResponseWriter, r *http.Request) {
	var res Response
	protocol, err := h.service.GetProtocolType(r.Context())
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = protocol
	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceHandler) ChartStream(w http.ResponseWriter, r *http.Request) {
	deviceId := r.URL.Query().Get("deviceId")
	if deviceId == "" {
		http.Error(w, "Missing deviceId", http.StatusBadRequest)
		return
	}

	// Set standard SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	rc := http.NewResponseController(w)

	_ = rc.Flush()
	clientChan := make(chan model.ChartData, 10)
	h.service.AddClient(deviceId, clientChan)
	defer h.service.RemoveClient(deviceId, clientChan)

	slog.InfoContext(r.Context(), "New Vue client connected to SSE stream!")

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(r.Context(), "Client disconnected from SSE stream")
			return

		case payload := <-clientChan:
			dataBytes, _ := json.Marshal(payload)
			fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))

			// ⚡ FIX: Ignore ErrNotSupported so logging/CORS middleware doesn't break the stream
			err := rc.Flush()
			if err != nil && !errors.Is(err, http.ErrNotSupported) {
				slog.ErrorContext(r.Context(), "Stream connection lost", slog.String("error", err.Error()))
				return
			}
		}
	}
}

func (h *DeviceHandler) GetAllDeviceName(w http.ResponseWriter, r *http.Request) {
	var res Response
	deviceDetails, err := h.service.GetAllDeviceName(r.Context())
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = deviceDetails
	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceHandler) ChartHistory(w http.ResponseWriter, r *http.Request) {
	var res Response
	var err error
	maxPointsStr := r.URL.Query().Get("maxPoints")
	deviceIdStr := r.URL.Query().Get("deviceIds")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	// layout := "2006-01-02T15:04:05"
	fromTime, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		res.Message = "Missing From Date"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}
	toTime, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		res.Message = "Missing To Date"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}
	deviceIds := []int{}
	for idStr := range strings.SplitSeq(deviceIdStr, ",") {
		if idInt, err := strconv.Atoi(idStr); err == nil {
			deviceIds = append(deviceIds, idInt)
		}
	}
	maxPoints, err := strconv.Atoi(maxPointsStr)
	if err != nil || maxPoints <= 0 {
		maxPoints = 100 // Default fallback
	}

	historyData, err := h.service.GetChartHistory(r.Context(), deviceIds, maxPoints, fromTime, toTime)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = historyData
	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceHandler) ValidateImport(w http.ResponseWriter, r *http.Request) {
	type importRow struct {
		DeviceName string  `json:"deviceName"`
		Protocol   *string `json:"protocol"`
		Active     bool    `json:"active"`
		IsValid    bool    `json:"isValid"`
		Message    string  `json:"message"`
	}

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
		MenuName:   "Device",
		ActionName: "Import",
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

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		res.Message = "File too large"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		res.Message = "No file uploaded"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))

	var rows []importRow
	const maxImportRows = 4000

	// 2. Parse based on file type
	switch ext {
	case ".json":
		// Matches the JSON export structure
		var incomingData []struct {
			DeviceName string `json:"deviceName"`
			Protocol   string `json:"protocol"`
			Status     string `json:"status"`
		}

		if err := json.NewDecoder(file).Decode(&incomingData); err != nil {
			res.Message = "Invalid JSON format"
			respondJson(w, http.StatusBadRequest, &res)
			return
		}

		if len(incomingData) > maxImportRows {
			res.Message = fmt.Sprintf("File exceeds maximum allowed rows (%d rows)", maxImportRows)
			respondJson(w, http.StatusBadRequest, &res)
			return
		}

		for _, d := range incomingData {
			active := true
			if strings.ToLower(strings.TrimSpace(d.Status)) == "inactive" {
				active = false
			}

			// ⚡ Handle Nullable Protocol
			var protocolPtr *string
			protoStr := strings.TrimSpace(d.Protocol)
			if protoStr != "" {
				protocolPtr = &protoStr
			}

			rows = append(rows, importRow{
				DeviceName: strings.TrimSpace(d.DeviceName),
				Protocol:   protocolPtr,
				Active:     active,
			})
		}
	case ".csv":
		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			res.Message = "Invalid CSV format"
			respondJson(w, http.StatusBadRequest, &res)
			return
		}

		if len(records) == 0 {
			res.Message = "Error: File is completely empty"
			respondJson(w, http.StatusBadRequest, &res)
			return
		}

		if len(records)-1 > maxImportRows {
			res.Message = fmt.Sprintf("File exceeds maximum allowed rows (%d rows)", maxImportRows)
			respondJson(w, http.StatusBadRequest, &res)
			return
		}

		// ⚡ STRICT HEADER VALIDATION
		header := records[0]
		if len(header) < 3 ||
			strings.ToLower(strings.TrimSpace(header[0])) != "devicename" ||
			strings.ToLower(strings.TrimSpace(header[1])) != "protocol" ||
			strings.ToLower(strings.TrimSpace(header[2])) != "status" {

			res.Message = "Invalid file template. Expected columns exactly as: 'deviceName', 'protocol', 'status'"
			respondJson(w, http.StatusBadRequest, &res)
			return
		}

		// Process rows (safely skipping the validated header)
		for i, record := range records {
			if i == 0 || len(record) == 0 {
				continue
			}

			deviceName := strings.TrimSpace(record[0])

			// ⚡ Handle Nullable Protocol safely
			var protocolPtr *string
			if len(record) >= 2 {
				protoStr := strings.TrimSpace(record[1])
				if protoStr != "" {
					protocolPtr = &protoStr
				}
			}

			active := true
			if len(record) >= 3 && strings.ToLower(strings.TrimSpace(record[2])) == "inactive" {
				active = false
			}

			rows = append(rows, importRow{
				DeviceName: deviceName,
				Protocol:   protocolPtr,
				Active:     active,
			})
		}

	case ".xlsx":
		f, err := excelize.OpenReader(file)
		if err != nil {
			res.Message = "Invalid Excel file"
			respondJson(w, http.StatusBadRequest, &res)
			return
		}
		defer f.Close()

		sheetMap := f.GetSheetMap()
		records, _ := f.GetRows(sheetMap[1])

		if len(records) == 0 {
			res.Message = "Error: Excel sheet is completely empty"
			respondJson(w, http.StatusBadRequest, &res)
			return
		}

		if len(records)-1 > maxImportRows {
			res.Message = fmt.Sprintf("File exceeds maximum allowed rows (%d rows)", maxImportRows)
			respondJson(w, http.StatusBadRequest, &res)
			return
		}

		// ⚡ STRICT HEADER VALIDATION
		header := records[0]
		if len(header) < 3 ||
			strings.ToLower(strings.TrimSpace(header[0])) != "devicename" ||
			strings.ToLower(strings.TrimSpace(header[1])) != "protocol" ||
			strings.ToLower(strings.TrimSpace(header[2])) != "status" {

			res.Message = "Invalid file template. Expected columns exactly as: 'deviceName', 'protocol', 'status'"
			respondJson(w, http.StatusBadRequest, &res)
			return
		}

		// Process rows (safely skipping the validated header)
		for i, record := range records {
			if i == 0 || len(record) == 0 {
				continue
			}

			deviceName := strings.TrimSpace(record[0])

			// ⚡ Handle Nullable Protocol safely
			var protocolPtr *string
			if len(record) >= 2 {
				protoStr := strings.TrimSpace(record[1])
				if protoStr != "" {
					protocolPtr = &protoStr
				}
			}

			active := true
			if len(record) >= 3 && strings.ToLower(strings.TrimSpace(record[2])) == "inactive" {
				active = false
			}

			rows = append(rows, importRow{
				DeviceName: deviceName,
				Protocol:   protocolPtr,
				Active:     active,
			})
		}

	default:
		res.Message = "Unsupported file type"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	// 3. Validation Logic
	protocolTypes, err := h.service.GetProtocolType(r.Context())
	if err != nil {
		res.Message = "Error fetching protocols"
		slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	validProtocol := make(map[string]bool, len(protocolTypes))
	for _, p := range protocolTypes {
		validProtocol[p] = true
	}

	devices, err := h.service.GetAllDeviceDetail(r.Context())
	if err != nil {
		res.Message = "Error fetching existing devices"
		slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	validDeviceName := make(map[string]bool, len(devices))
	for _, d := range devices {
		validDeviceName[d.DeviceName] = true
	}

	for i := range rows {
		rows[i].IsValid = true
		rows[i].Message = "Valid to Import"

		switch {
		case rows[i].DeviceName == "":
			rows[i].IsValid = false
			rows[i].Message = "Error: Device Name is required"

		// ⚡ ONLY check validProtocol if a protocol was actually provided
		case rows[i].Protocol != nil && !validProtocol[strings.ToUpper(*rows[i].Protocol)]:
			rows[i].IsValid = false
			rows[i].Message = "Error: Unsupported Protocol type"

		case validDeviceName[rows[i].DeviceName]:
			rows[i].IsValid = false
			rows[i].Message = "Error: Device is Duplicate"
		}
	}

	res.Message = "Validation complete"
	res.Data = rows
	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceHandler) ExportDevices(w http.ResponseWriter, r *http.Request) {
	var res Response
	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Device",
		ActionName: "Export",
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
	format := r.URL.Query().Get("format")

	// Fetch actual data from your database repository
	devices, err := h.service.GetAllDeviceDetail(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}

	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment;filename=devices.csv")
		writer := csv.NewWriter(w)
		writer.Write([]string{"deviceName", "protocol", "status"}) // Headers
		for _, dev := range devices {
			statusStr := "inactive"
			if dev.Active {
				statusStr = "active"
			}

			// ⚡ Safe pointer dereference
			protocolStr := ""
			if dev.Protocol != nil {
				protocolStr = *dev.Protocol
			}

			writer.Write([]string{dev.DeviceName, protocolStr, statusStr})
		}
		writer.Flush()

	case "json":
		var exportData []struct {
			DeviceName string `json:"deviceName"`
			Protocol   string `json:"protocol"`
			Status     string `json:"status"`
		}

		for _, d := range devices {
			statusStr := "inactive"
			if d.Active {
				statusStr = "active"
			}

			// ⚡ Safe pointer dereference
			protocolStr := ""
			if d.Protocol != nil {
				protocolStr = *d.Protocol
			}

			exportData = append(exportData, struct {
				DeviceName string `json:"deviceName"`
				Protocol   string `json:"protocol"`
				Status     string `json:"status"`
			}{
				DeviceName: d.DeviceName,
				Protocol:   protocolStr,
				Status:     statusStr,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment;filename=devices.json")
		json.NewEncoder(w).Encode(exportData)

	case "excel":
		f := excelize.NewFile()
		f.SetCellValue("Sheet1", "A1", "deviceName")
		f.SetCellValue("Sheet1", "B1", "protocol")
		f.SetCellValue("Sheet1", "C1", "status")

		for i, dev := range devices {
			rowNum := i + 2
			statusStr := "inactive"
			if dev.Active {
				statusStr = "active"
			}

			// ⚡ Safe pointer dereference
			protocolStr := ""
			if dev.Protocol != nil {
				protocolStr = *dev.Protocol
			}

			f.SetCellValue("Sheet1", fmt.Sprintf("A%d", rowNum), dev.DeviceName)
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", rowNum), protocolStr)
			f.SetCellValue("Sheet1", fmt.Sprintf("C%d", rowNum), statusStr)
		}

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment;filename=devices.xlsx")
		f.Write(w)
	}
}

func (h *DeviceHandler) GetAllGroupDetail(w http.ResponseWriter, r *http.Request) {
	var res Response
	deviceDetails, err := h.service.GetDeviceGroupDetail(r.Context())
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = deviceDetails
	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var res Response
	var err error
	var deviceGroup model.CreateDeviceGroup

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Device Group",
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

	if err = json.NewDecoder(r.Body).Decode(&deviceGroup); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	err = h.service.CreateDeviceGroup(r.Context(), deviceGroup, authUserId)
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

func (h *DeviceHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	var res Response
	var err error
	var deviceGroup model.UpdateDeviceGroup

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Device Group",
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

	if err = json.NewDecoder(r.Body).Decode(&deviceGroup); err != nil {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	err = h.service.UpdateDeviceGroup(r.Context(), &deviceGroup, authUserId)
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

	go func(oldName string) {
		// Use a fresh context since the request context will be cancelled immediately
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.deviceGatewayClient.InvalidateGroupCache(bgCtx, oldName); err != nil {
			slog.Error("Failed to invalidate gateway cache", slog.String("error", err.Error()))
		}
	}(deviceGroup.OldName)

	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
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
		MenuName:   "Device Group",
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
	deleteDeviceGroupId, err := strconv.Atoi(idStr)
	if err != nil {
		res.Message = "Invalid group Id"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	deviceGroup, err := h.service.DeleteDeviceGroup(r.Context(), deleteDeviceGroupId, authUserId)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if deviceGroup != nil {
		go func(oldName string) {
			// Use a fresh context since the request context will be cancelled immediately
			bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := h.deviceGatewayClient.InvalidateGroupCache(bgCtx, oldName); err != nil {
				slog.Error("Failed to invalidate gateway cache", slog.String("error", err.Error()))
			}
		}(deviceGroup.GroupName)
	}

	respondJson(w, http.StatusOK, &res)
}
