package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strconv"

	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type DeviceGatewayHandler struct {
	service        model.DeviceGatewayService
	sessionService model.SessionManagerService
	cacheService   model.DeviceCacheService
}

func NewDeviceGatewayHandler(gatewayService model.DeviceGatewayService, sessionService model.SessionManagerService, cacheService model.DeviceCacheService) *DeviceGatewayHandler {
	return &DeviceGatewayHandler{
		service:        gatewayService,
		sessionService: sessionService,
		cacheService:   cacheService,
	}
}

func (h *DeviceGatewayHandler) Command(w http.ResponseWriter, r *http.Request) {
	var res Response
	var req model.GatewayCommand

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		res.Message = "Invalid request body"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	hasDevice := req.DeviceId > 0
	hasGroup := req.GroupId > 0

	// ⚡ FIX: Changed len(req.Payload) != 0 to == 0
	// It ensures you provide either a DeviceId OR a GroupId (not both),
	// ensures the payload has items, and ensures a protocol exists.
	if (hasDevice == hasGroup) || len(req.Payload) == 0 || req.Protocol == "" {
		res.Message = "Missing required fields (Requires DeviceId XOR GroupId, valid Payload, and Protocol)"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	detachedCtx := context.WithoutCancel(r.Context())

	// Delegate the actual routing to the SessionManager Service
	err := h.sessionService.RouteCommand(detachedCtx, &req)
	if err != nil {
		res.Message = "Error routing command"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
			slog.Int("deviceId", req.DeviceId),
			slog.Int("groupId", req.GroupId),
			slog.String("protocol", req.Protocol),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Message = "Command routed successfully"
	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceGatewayHandler) HTTPTelemetry(w http.ResponseWriter, r *http.Request) {
	var res Response
	var data []model.DeviceDataPayloadReq

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		res.Message = "Failed to read request body"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}
	defer r.Body.Close()

	packetData := bytes.TrimSpace(bodyBytes)
	if len(packetData) == 0 {
		res.Message = "Empty request body"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	switch packetData[0] {
	case '[':
		if err := json.Unmarshal(packetData, &data); err != nil {
			res.Message = "Invalid JSON array"
			respondJson(w, http.StatusBadRequest, &res)
			return
		}
	case '{':
		var single model.DeviceDataPayloadReq
		if err := json.Unmarshal(packetData, &single); err != nil {
			res.Message = "Invalid JSON object"
			respondJson(w, http.StatusBadRequest, &res)
			return
		}
		data = append(data, single)
	default:
		res.Message = "Invalid JSON format"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	ctx := r.Context()

	groupName := r.URL.Query().Get("device-group")
	targetDevice := r.URL.Query().Get("device") // ⚡ NEW: Get the target device from the URL

	validGroupDevices := make(map[int]bool)
	isGroupMode := groupName != ""

	if isGroupMode {
		deviceIDs, groupProtocol, err := h.cacheService.GetGroupInfoByName(ctx, groupName)
		if err != nil || len(deviceIDs) == 0 {
			res.Message = "Group not found or empty"
			respondJson(w, http.StatusNotFound, &res)
			return
		}
		if groupProtocol != "HTTP" {
			res.Message = "Group is not configured for HTTP protocol"
			respondJson(w, http.StatusBadRequest, &res)
			return
		}
		for _, id := range deviceIDs {
			validGroupDevices[id] = true
		}
	}

	successCount := 0

	for _, reqData := range data {
		if reqData.DeviceName == "" {
			continue
		}

		// ⚡ NEW: If not in group mode, enforce the ?device= parameter (if provided)
		if !isGroupMode && targetDevice != "" && reqData.DeviceName != targetDevice {
			slog.DebugContext(ctx, "Skipped device (does not match URL target)",
				slog.String("payloadDevice", reqData.DeviceName),
				slog.String("urlTarget", targetDevice),
			)
			continue
		}

		deviceId, deviceProtocol, err := h.cacheService.GetDeviceInfoByName(ctx, reqData.DeviceName)
		if err != nil || deviceId <= 0 {
			slog.WarnContext(ctx, "Unknown device name in HTTP Telemetry", slog.String("name", reqData.DeviceName))
			continue
		}

		if isGroupMode {
			if !validGroupDevices[deviceId] {
				slog.WarnContext(ctx, "Device does not belong to the specified group", slog.String("device", reqData.DeviceName))
				continue
			}
		} else {
			if deviceProtocol != "HTTP" {
				slog.WarnContext(ctx, "Device is not configured for HTTP", slog.String("device", reqData.DeviceName))
				continue
			}
		}

		deviceData := model.DeviceData{
			DeviceId:   deviceId,
			DeviceName: reqData.DeviceName,
			ValueData:  int(math.Round(reqData.ValueData * model.DeviceScale)),
		}

		h.sessionService.MarkDeviceActive(deviceId)
		h.service.Add(deviceData)
		successCount++
	}

	if successCount == 0 && len(data) > 0 {
		res.Message = "No valid devices processed from payload"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	res.Message = "Success"
	respondJson(w, http.StatusOK, &res)
}

// HTTPCommandPolling allows HTTP physical devices to fetch queued commands via GET
func (h *DeviceGatewayHandler) HTTPCommandPolling(w http.ResponseWriter, r *http.Request) {
	var res Response
	var allCommands []model.DeviceCommand

	ctx := r.Context()
	groupName := r.URL.Query().Get("device-group")

	if groupName != "" {
		// --- NEW: Validate Group Protocol ---
		deviceIds, groupProtocol, err := h.cacheService.GetGroupInfoByName(ctx, groupName)
		if err != nil || len(deviceIds) == 0 {
			res.Message = "Group not found or contains no devices"
			respondJson(w, http.StatusNotFound, &res)
			return
		}

		if groupProtocol != "HTTP" {
			res.Message = "Group is not configured for HTTP protocol"
			respondJson(w, http.StatusBadRequest, &res)
			return
		}
		// ------------------------------------

		for _, id := range deviceIds {
			h.sessionService.MarkDeviceActive(id)
			if payload, exists := h.sessionService.PopHTTPCommand(id); exists {
				allCommands = append(allCommands, payload...)
			}
		}

	} else {
		deviceName := r.URL.Query().Get("device")
		if deviceName == "" {
			res.Message = "Missing 'device' or 'device-group' parameter"
			respondJson(w, http.StatusBadRequest, &res)
			return
		}

		deviceId, protocol, err := h.cacheService.GetDeviceInfoByName(ctx, deviceName)
		if err != nil || deviceId <= 0 {
			res.Message = "Device not found"
			respondJson(w, http.StatusNotFound, &res)
			return
		}

		if protocol != "HTTP" {
			res.Message = "Device is not configured for HTTP protocol"
			respondJson(w, http.StatusBadRequest, &res)
			return
		}

		h.sessionService.MarkDeviceActive(deviceId)
		if payload, exists := h.sessionService.PopHTTPCommand(deviceId); exists {
			allCommands = append(allCommands, payload...)
		}
	}

	if len(allCommands) == 0 {
		respondJson(w, http.StatusNoContent, nil)
		return
	}

	res.Data = allCommands
	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceGatewayHandler) DeviceStatus(w http.ResponseWriter, r *http.Request) {
	var res Response // Using your standard response struct

	deviceIdStr := r.URL.Query().Get("deviceId")
	deviceId, err := strconv.Atoi(deviceIdStr)
	if err != nil || deviceId <= 0 {
		res.Message = "Invalid deviceId parameter"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	// Ask the Session Manager!
	isOnline := h.sessionService.IsDeviceOnline(deviceId)
	if isOnline {
		if err = h.service.UpdateDeviceLastSeen(r.Context(), deviceId); err != nil {
			slog.Error("Error",
				slog.String("track", err.Error()),
			)
		}
	}

	res.Message = "Success"
	res.Data = map[string]any{
		"deviceId": deviceId,
		"isOnline": isOnline,
	}

	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceGatewayHandler) ClearCache(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	cacheType := r.URL.Query().Get("type") // expecting "device" or "group"

	if name == "" {
		http.Error(w, "Missing 'name' parameter", http.StatusBadRequest)
		return
	}

	switch cacheType {
	case "device":
		// ⚡ Use the injected Cache Service instead of the listener package
		h.cacheService.InvalidateDevice(name)
		slog.Info("Invalidated device cache", slog.String("deviceName", name))
	case "group":
		// ⚡ Use the injected Cache Service
		h.cacheService.InvalidateGroup(name)
		slog.Info("Invalidated group cache", slog.String("groupName", name))
	default:
		http.Error(w, "Invalid 'type' parameter. Must be 'device' or 'group'", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
