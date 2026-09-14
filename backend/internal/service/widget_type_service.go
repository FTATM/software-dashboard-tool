package service

import (
	"context"
	"fmt"

	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type widgetTypeService struct {
	txManager      model.TransactionManager
	prefixError    string
	widgetTypeRepo model.WidgetTypeRepository
}

func NewWidgetTypeService(txManager model.TransactionManager, wrt model.WidgetTypeRepository) model.WidgetTypeService {
	return &widgetTypeService{txManager: txManager, prefixError: "widgetTypeService", widgetTypeRepo: wrt}
}

func (s *widgetTypeService) GetWidgetTypeById(ctx context.Context, canvasId int) (*model.WidgetType, error) {
	const fname = "GetWidgetTypeById"
	widgetType, err := s.widgetTypeRepo.GetById(ctx, canvasId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return widgetType, nil
}

func (s *widgetTypeService) GetWidgetTypeAll(ctx context.Context) ([]model.WidgetType, error) {
	const fname = "GetWidgetTypeAll"
	widgetTypes, err := s.widgetTypeRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return widgetTypes, nil
}
