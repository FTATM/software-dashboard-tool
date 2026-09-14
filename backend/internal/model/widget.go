package model

import (
	"context"
)

type Widget struct {
	WidgetId        int         `json:"widgetId" db:"widget_id"`
	WidgetTypeId    int         `json:"widgetTypeId" db:"widget_type_id"`
	CanvasId        int         `json:"canvasId" db:"canvas_id"`
	DeviceGroupId   *int        `json:"deviceGroupId" db:"device_group_id"`
	DeviceIds       []int       `json:"deviceIds" db:"device_id_s"`
	WidgetLabel     string      `json:"widgetLabel" db:"widget_label"`
	LayoutData      DynamicJSON `json:"layoutData" db:"layout_data"`
	WidgetStyle     DynamicJSON `json:"widgetStyle" db:"widget_style"`
	CustomChartData DynamicJSON `json:"customChartData" db:"custom_chart_data"`
}

type UpsertWidgetReqest struct {
	CanvasId      int            `json:"canvasId"`
	CanvasStyle   DynamicJSON    `json:"canvasStyle,omitempty"`
	UpsertWidgets []UpsertWidget `json:"upsertWidgets"`
}

type UpsertWidget struct {
	WidgetId        int         `json:"widgetId"`
	WidgetTypeId    int         `json:"widgetTypeId"`
	DeviceGroupId   *int        `json:"deviceGroupId"`
	DeviceIds       []int       `json:"deviceIds"`
	WidgetLabel     string      `json:"widgetLabel"`
	LayoutData      DynamicJSON `json:"layoutData"`
	WidgetStyle     DynamicJSON `json:"widgetStyle"`
	CustomChartData DynamicJSON `json:"customChartData"`
}

type WidgetRepository interface {
	GetById(ctx context.Context, id int) (*Widget, error)
	Create(ctx context.Context, widgets []Widget) error
	Update(ctx context.Context, widgets []Widget) error
	Delete(ctx context.Context, ids []int) error
	GetWidgetByCanvasId(ctx context.Context, canvasId []int) ([]Widget, error)
	GetActiveImageFilenames(ctx context.Context) ([]string, error)
}

type WidgetService interface {
	GetWidgetDetailById(ctx context.Context, widgetId int) (*Widget, error)
	CreateWidget(ctx context.Context, widgets []Widget) error
	UpdateWidget(ctx context.Context, widgets []Widget) error
	DeleteWidget(ctx context.Context, widgetIds []int) error
	UpsertWidget(ctx context.Context, req *UpsertWidgetReqest, authUserId int) error
}
