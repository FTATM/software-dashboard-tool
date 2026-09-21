package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/jackc/pgx/v5"
)

type canvasService struct {
	txManager    model.TransactionManager
	prefixError  string
	canvasRepo   model.CanvasRepository
	widgetRepo   model.WidgetRepository
	auditLogRepo model.AuditLogRepository
}

func NewCanvasService(txManager model.TransactionManager, wr model.WidgetRepository, cr model.CanvasRepository, auditlogRepo model.AuditLogRepository) model.CanvasService {
	return &canvasService{txManager: txManager, prefixError: "canvasService", widgetRepo: wr, canvasRepo: cr, auditLogRepo: auditlogRepo}
}

func (s *canvasService) GetAllCanvas(ctx context.Context) ([]model.Canvas, error) {
	const fname = "GetAllCanvasByUserRole"

	allCanvas, err := s.canvasRepo.GetAll(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return allCanvas, nil
}

func (s *canvasService) GetCanvasDetailById(ctx context.Context, canvasId int) (*model.CanvasDetail, error) {
	const fname = "GetCanvasDetailById"

	canvas, err := s.canvasRepo.GetById(ctx, canvasId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	widgets, err := s.widgetRepo.GetWidgetByCanvasId(ctx, []int{canvas.CanvasId})
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return &model.CanvasDetail{
		CanvasId:    canvas.CanvasId,
		CanvasName:  canvas.CanvasName,
		CanvasStyle: canvas.CanvasStyle,
		Widgets:     widgets,
	}, nil
}

func (s *canvasService) GetAllCanvasDetailByUserRole(ctx context.Context, authUserId int) ([]model.CanvasDetail, error) {
	const fname = "GetAllCanvasDetailByUserRole"

	userCanvas, err := s.canvasRepo.GetCanvasByUserId(ctx, authUserId, true)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	canvasIds := make([]int, 0, len(userCanvas))
	for _, canvas := range userCanvas {
		canvasIds = append(canvasIds, canvas.CanvasId)
	}

	widgets, err := s.widgetRepo.GetWidgetByCanvasId(ctx, canvasIds)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	widgetMap := make(map[int][]model.Widget, len(userCanvas))
	for _, widget := range widgets {
		widgetMap[widget.CanvasId] = append(widgetMap[widget.CanvasId], widget)
	}

	result := make([]model.CanvasDetail, 0, len(userCanvas))

	for _, canvas := range userCanvas {
		detail := model.CanvasDetail{
			CanvasId:    canvas.CanvasId,
			CanvasName:  canvas.CanvasName,
			CanvasStyle: canvas.CanvasStyle,
			Widgets:     widgetMap[canvas.CanvasId],
		}
		result = append(result, detail)
	}

	return result, nil
}

func (s *canvasService) GetAllCanvasRoleDetail(ctx context.Context) ([]model.CanvasRoleDetail, error) {
	const fname = "GetAllCanvasRoleDetail"

	allCanvasRole, err := s.canvasRepo.GetAllCanvasRole(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	roleMap := make(map[int][]int)
	for _, cr := range allCanvasRole {
		roleMap[cr.RoleId] = append(roleMap[cr.RoleId], cr.CanvasId)
	}

	var detail []model.CanvasRoleDetail
	for roleId, canvasIds := range roleMap {
		detail = append(detail, model.CanvasRoleDetail{
			RoleId:    roleId,
			CanvasIds: canvasIds,
		})
	}

	return detail, nil
}

func (s *canvasService) UpsertCanvasRole(ctx context.Context, upsertCanvasRole *model.UpsertCanvasRole, authUserId int) error {
	const fname = "UpsertCanvasRole"
	var err error

	oldCanvasIds, err := s.canvasRepo.GetCanvasRoleByRoleId(ctx, upsertCanvasRole.RoleId)
	oldCanvasMap := make(map[int]bool)
	for _, oc := range oldCanvasIds {
		oldCanvasMap[oc] = true
	}

	var createCanvasRole, deleteCanvasRole []model.CanvasRole

	for _, cId := range upsertCanvasRole.CanvasIds {
		if _, found := oldCanvasMap[cId]; found {
			// Found in both!
			// Delete it from the map so we know it's been handled.
			delete(oldCanvasMap, cId)
		} else {
			//Not found in old map! This is a brand new ID.
			createCanvasRole = append(createCanvasRole, model.CanvasRole{RoleId: upsertCanvasRole.RoleId, CanvasId: cId})
		}
	}

	for cId := range oldCanvasMap {
		deleteCanvasRole = append(deleteCanvasRole, model.CanvasRole{RoleId: upsertCanvasRole.RoleId, CanvasId: cId})
	}

	if len(createCanvasRole) == 0 && len(deleteCanvasRole) == 0 {
		return nil
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	err = s.canvasRepo.CreateCanvasRole(ctx, createCanvasRole)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	err = s.canvasRepo.DeleteCanvasRole(ctx, deleteCanvasRole)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	newData, err := model.StructToDynamicJSON(upsertCanvasRole)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "canvas_role",
		EntityId:   strconv.Itoa(upsertCanvasRole.RoleId),
		MenuType:   "Canvas Access",
		Action:     model.UpdateAction,
		ChangedBy:  authUserId,
		OldData:    nil,
		NewData:    newData,
	}

	if err = s.auditLogRepo.Create(tx.Context(), []model.AuditLog{audit}); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *canvasService) CreateCanvas(ctx context.Context, createCanvas *model.CreateCanvas, authUserId int) error {
	const fname = "CreateCanvas"

	canvas := model.Canvas{
		CanvasName: createCanvas.CanvasName,
	}

	duplicate, err := s.canvasRepo.CountValidate(ctx, &canvas)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if duplicate > 0 {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, model.ErrDuplicate)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	err = s.canvasRepo.Create(ctx, &canvas)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, 1)
	newData, err := model.StructToDynamicJSON(canvas)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "canvas",
		EntityId:   strconv.Itoa(canvas.CanvasId),
		MenuType:   "Canvas",
		Action:     model.CreateAction,
		ChangedBy:  authUserId,
		OldData:    nil,
		NewData:    newData,
	}
	auditlogs = append(auditlogs, audit)

	if err = s.auditLogRepo.Create(tx.Context(), auditlogs); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}
func (s *canvasService) UpdateCanvas(ctx context.Context, updateCanvas *model.UpdateCanvas, authUserId int) error {
	const fname = "UpdateCanvas"

	canvas := model.Canvas{
		CanvasId:   updateCanvas.CanvasId,
		CanvasName: updateCanvas.CanvasName,
	}

	duplicate, err := s.canvasRepo.CountValidate(ctx, &canvas)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if duplicate > 0 {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, model.ErrDuplicate)
	}

	oldCanvas, err := s.canvasRepo.GetById(ctx, canvas.CanvasId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if oldCanvas.IsSame(canvas) {
		return nil
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	err = s.canvasRepo.Update(ctx, &canvas)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, 1)
	oldData, err := model.StructToDynamicJSON(oldCanvas)
	newData, err := model.StructToDynamicJSON(canvas)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "canvas",
		EntityId:   strconv.Itoa(canvas.CanvasId),
		MenuType:   "Canvas",
		Action:     model.UpdateAction,
		ChangedBy:  authUserId,
		OldData:    oldData,
		NewData:    newData,
	}
	auditlogs = append(auditlogs, audit)

	if err = s.auditLogRepo.Create(tx.Context(), auditlogs); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *canvasService) DeleteCanvas(ctx context.Context, canvasId int, authUserId int) error {
	const fname = "DeleteCanvas"

	oldCanvas, err := s.canvasRepo.GetById(ctx, canvasId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	err = s.canvasRepo.Delete(ctx, canvasId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	auditlogs := make([]model.AuditLog, 0, 1)
	oldData, err := model.StructToDynamicJSON(oldCanvas)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	audit := model.AuditLog{
		EntityType: "canvas",
		EntityId:   strconv.Itoa(canvasId),
		MenuType:   "Canvas",
		Action:     model.DeleteAction,
		ChangedBy:  authUserId,
		OldData:    oldData,
		NewData:    nil,
	}
	auditlogs = append(auditlogs, audit)

	if err = s.auditLogRepo.Create(tx.Context(), auditlogs); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *canvasService) ExecuteDynamicQuery(ctx context.Context, rawQuery string, authUserId int) ([]map[string]any, error) {
	const fname = "ExecuteDynamicQuery"

	cleanQuery := strings.TrimSpace(rawQuery)
	upperQuery := strings.ToUpper(strings.TrimSpace(cleanQuery))

	if !strings.HasPrefix(upperQuery, "SELECT") && !strings.HasPrefix(upperQuery, "WITH") {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, model.ErrSecurityViolation)
	}

	forbiddenKeywords := []string{
		"PG_",                // Blocks all pg_catalog tables (pg_user, pg_class, etc.)
		"INFORMATION_SCHEMA", // Blocks schema exploration
		"CURRENT_USER",       // Blocks session details
		"SESSION_USER",
		"CURRENT_DATABASE",
		"VERSION(",
		"_TIMESCALEDB",
		"TIMESCALEDB_INFORMATION",
	}

	for _, keyword := range forbiddenKeywords {
		if strings.Contains(upperQuery, keyword) {
			return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, model.ErrSecurityViolation)
		}
	}

	cleanQuery = strings.TrimRight(cleanQuery, "; \t\n")
	safeQuery := fmt.Sprintf("SELECT * FROM (%s) AS user_query LIMIT 500", cleanQuery)
	readTx, err := s.txManager.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	results, err := s.canvasRepo.ExecuteDynamicQuery(readTx.Context(), safeQuery)
	readTx.Rollback(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	// ⚡ --- NEW: HARDCODED COLUMN FILTER --- ⚡
	restrictedColumns := []string{
		"password",
		"password_hash",
		"token",
		"line_user_token",
		"refresh_token",
		"api_key",
		"secret",
	}

	for _, row := range results {
		for _, col := range restrictedColumns {
			delete(row, col)
		}

		for key, val := range row {
			// pgx/v5 returns dynamic UUIDs as [16]byte
			if b, ok := val.([16]byte); ok {
				// Convert the raw bytes into the standard 8-4-4-4-12 UUID string format
				row[key] = fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
			}

			// Optional: If you have text/varchar columns coming back as raw []byte,
			// this ensures they render as strings instead of Base64 gibberish
			if b, ok := val.([]byte); ok {
				row[key] = string(b)
			}
		}
	}

	writeTx, err := s.txManager.Begin(ctx)
	if err == nil {
		defer writeTx.Rollback(ctx)

		// Store the exact string they typed into the JSON log
		queryPayload := map[string]string{"raw_query": rawQuery}
		newData, err := model.StructToDynamicJSON(queryPayload)
		if err != nil {
			return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}

		audit := model.AuditLog{
			EntityType: "dynamic_query",
			EntityId:   "0",
			MenuType:   "Dashboard",
			Action:     model.QueryAction,
			ChangedBy:  authUserId,
			OldData:    nil,
			NewData:    newData,
		}

		if err := s.auditLogRepo.Create(writeTx.Context(), []model.AuditLog{audit}); err != nil {
			return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		} else {
			_ = writeTx.Commit(ctx)
		}
	}
	return results, nil
}
