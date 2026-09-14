package model

import (
	"context"
	"time"
)

type WidgetType struct {
	WidgetTypeId   int       `json:"widgetTypeId" db:"widget_type_id"`
	WidgetTypeName string    `json:"widgetTypeName" db:"widget_type_name"`
	CreatedAt      time.Time `json:"-" db:"created_at"`
}

type WidgetTypeRepository interface {
	GetById(ctx context.Context, id int) (*WidgetType, error)
	GetAll(ctx context.Context) ([]WidgetType, error)
}

type WidgetTypeService interface {
	GetWidgetTypeById(ctx context.Context, id int) (*WidgetType, error)
	GetWidgetTypeAll(ctx context.Context) ([]WidgetType, error)
}
