package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type widgetService struct {
	txManager      model.TransactionManager
	prefixError    string
	widgetRepo     model.WidgetRepository
	widgetTypeRepo model.WidgetTypeRepository
	canvasRepo     model.CanvasRepository
	auditLogRepo   model.AuditLogRepository
}

func NewWidgetService(txManager model.TransactionManager, wr model.WidgetRepository, wrt model.WidgetTypeRepository, cr model.CanvasRepository, auditLogRepo model.AuditLogRepository) model.WidgetService {
	return &widgetService{
		txManager:      txManager,
		widgetRepo:     wr,
		widgetTypeRepo: wrt,
		canvasRepo:     cr,
		prefixError:    "widgetService",
		auditLogRepo:   auditLogRepo,
	}
}

func (s *widgetService) GetWidgetDetailById(ctx context.Context, widgetId int) (*model.Widget, error) {
	const fname = "GetWidgetDetailById"
	widget, err := s.widgetRepo.GetById(ctx, widgetId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return &model.Widget{
		WidgetId:     widget.WidgetId,
		WidgetTypeId: widget.WidgetTypeId,
		LayoutData:   widget.LayoutData,
	}, nil
}

func (s *widgetService) CreateWidget(ctx context.Context, widgets []model.Widget) error {
	const fname = "CreateWidget"
	var err error

	if len(widgets) == 0 {
		return nil
	}

	_, err = s.canvasRepo.GetById(ctx, widgets[0].CanvasId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	widgetTypes, err := s.widgetTypeRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	widgetTypeMap := make(map[int]string)
	for _, widgetType := range widgetTypes {
		widgetTypeMap[widgetType.WidgetTypeId] = widgetType.WidgetTypeName
	}

	for _, widget := range widgets {
		_, exists := widgetTypeMap[widget.WidgetTypeId]

		if !exists {
			return fmt.Errorf("[%s]>[%s]: invalid widget type Id provided: %d", s.prefixError, fname, widget.WidgetTypeId)
		}

	}

	if err = s.widgetRepo.Create(ctx, widgets); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *widgetService) UpdateWidget(ctx context.Context, widgets []model.Widget) error {
	const fname = "UpdateWidget"
	if len(widgets) == 0 {
		return nil
	}

	if err := s.widgetRepo.Update(ctx, widgets); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *widgetService) DeleteWidget(ctx context.Context, widgetIds []int) error {
	const fname = "DeleteWidget"
	if len(widgetIds) == 0 {
		return nil
	}
	if err := s.widgetRepo.Delete(ctx, widgetIds); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *widgetService) UpsertWidget(ctx context.Context, upsertWidget *model.UpsertWidgetReqest, authUserId int) error {
	const fname = "UpsertWidget"
	var err error
	var toCreate, toUpdate []model.Widget
	var toDelete []int

	dbWidgets, err := s.widgetRepo.GetWidgetByCanvasId(ctx, []int{upsertWidget.CanvasId})
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if len(dbWidgets) == 0 && len(upsertWidget.UpsertWidgets) == 0 {
		return nil
	}

	reqMap := make(map[int]model.Widget)
	for _, reqWidget := range upsertWidget.UpsertWidgets {
		if reqWidget.WidgetId == 0 {
			tempCreate := model.Widget{
				WidgetTypeId:    reqWidget.WidgetTypeId,
				CanvasId:        upsertWidget.CanvasId,
				WidgetLabel:     reqWidget.WidgetLabel,
				DeviceGroupId:   reqWidget.DeviceGroupId,
				DeviceIds:       reqWidget.DeviceIds,
				LayoutData:      reqWidget.LayoutData,
				WidgetStyle:     reqWidget.WidgetStyle,
				CustomChartData: reqWidget.CustomChartData,
			}
			toCreate = append(toCreate, tempCreate)
		} else {
			tempUpdate := model.Widget{
				WidgetId:        reqWidget.WidgetId,
				WidgetTypeId:    reqWidget.WidgetTypeId,
				CanvasId:        upsertWidget.CanvasId,
				WidgetLabel:     reqWidget.WidgetLabel,
				DeviceGroupId:   reqWidget.DeviceGroupId,
				DeviceIds:       reqWidget.DeviceIds,
				LayoutData:      reqWidget.LayoutData,
				WidgetStyle:     reqWidget.WidgetStyle,
				CustomChartData: reqWidget.CustomChartData,
			}
			reqMap[reqWidget.WidgetId] = tempUpdate
		}
	}

	for _, dbWidget := range dbWidgets {
		if reqItem, exists := reqMap[dbWidget.WidgetId]; exists {
			toUpdate = append(toUpdate, reqItem)
		} else {
			toDelete = append(toDelete, dbWidget.WidgetId)
		}
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	if upsertWidget.CanvasStyle != nil {
		oldCanvas, err := s.canvasRepo.GetById(ctx, upsertWidget.CanvasId)
		if err != nil {
			return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}

		canvas := model.Canvas{CanvasId: upsertWidget.CanvasId, CanvasName: oldCanvas.CanvasName, CanvasStyle: upsertWidget.CanvasStyle}
		if !oldCanvas.IsSame(canvas) {
			if err = s.canvasRepo.Update(tx.Context(), &canvas); err != nil {
				return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
			}
		}
	}

	if err = s.DeleteWidget(tx.Context(), toDelete); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	if err = s.CreateWidget(tx.Context(), toCreate); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err = s.UpdateWidget(tx.Context(), toUpdate); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	audit := model.AuditLog{
		EntityType: "canvas",
		EntityId:   strconv.Itoa(upsertWidget.CanvasId),
		MenuType:   "Canvas Design",
		Action:     model.UpdateAction,
		ChangedBy:  authUserId,
		OldData:    nil,
		NewData:    nil,
	}

	if err = s.auditLogRepo.Create(tx.Context(), []model.AuditLog{audit}); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}
