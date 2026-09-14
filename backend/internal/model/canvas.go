package model

import (
	"context"
	"time"
)

type Canvas struct {
	CanvasId    int         `json:"canvasId" db:"canvas_id"`
	CanvasName  string      `json:"canvasName" db:"canvas_name"`
	CanvasStyle DynamicJSON `json:"canvasStyle" db:"canvas_style"`
	CreatedAt   time.Time   `json:"-" db:"created_at"`
}

func (s *Canvas) IsSame(req Canvas) bool {
	if s == nil {
		return false
	}

	return s.CanvasId == req.CanvasId &&
		s.CanvasName == req.CanvasName &&
		s.CanvasStyle.AreJSONsEqual(req.CanvasStyle)
}

type CanvasDetail struct {
	CanvasId    int         `json:"canvasId"`
	CanvasName  string      `json:"canvasName"`
	CanvasStyle DynamicJSON `json:"canvasStyle"`
	Widgets     []Widget    `json:"widgets"`
}

type UpdateCanvas struct {
	CanvasId    int         `json:"canvasId"`
	CanvasName  string      `json:"canvasName"`
	CanvasStyle DynamicJSON `json:"canvasStyle,omitempty"`
}

type UpsertCanvasRole struct {
	RoleId    int   `json:"roleId"`
	CanvasIds []int `json:"canvasIds"`
}

type CreateCanvas struct {
	CanvasName  string      `json:"canvasName"`
	CanvasStyle DynamicJSON `json:"canvasStyle,omitempty"`
}

type CanvasRepository interface {
	GetAll(ctx context.Context, active bool) ([]Canvas, error)
	GetById(ctx context.Context, id int) (*Canvas, error)
	GetByIds(ctx context.Context, ids []int) ([]Canvas, error)
	GetCanvasByUserId(ctx context.Context, userId int, active bool) ([]Canvas, error)
	GetCanvasByRoleId(ctx context.Context, roleId int, active bool) ([]Canvas, error)
	GetAllCanvasRole(ctx context.Context) ([]CanvasRole, error)
	GetCanvasRoleByRoleId(ctx context.Context, roleId int) ([]int, error)
	CreateCanvasRole(ctx context.Context, canvasroles []CanvasRole) error
	DeleteCanvasRole(ctx context.Context, canvasroles []CanvasRole) error
	Create(ctx context.Context, canvas *Canvas) error
	Update(ctx context.Context, canvas *Canvas) error
	Delete(ctx context.Context, canvasId int) error
	CountValidate(ctx context.Context, canvas *Canvas) (int, error)
	ExecuteDynamicQuery(ctx context.Context, rawQuery string) ([]map[string]any, error)
}

type CanvasService interface {
	GetAllCanvas(ctx context.Context) ([]Canvas, error)
	GetCanvasDetailById(ctx context.Context, id int) (*CanvasDetail, error)
	GetAllCanvasDetailByUserRole(ctx context.Context, authUserId int) ([]CanvasDetail, error)
	GetAllCanvasRoleDetail(ctx context.Context) ([]CanvasRoleDetail, error)
	UpsertCanvasRole(ctx context.Context, upsertCanvasRole *UpsertCanvasRole, authUserId int) error
	CreateCanvas(ctx context.Context, createCanvas *CreateCanvas, authUserId int) error
	UpdateCanvas(ctx context.Context, updateCanvas *UpdateCanvas, authUserId int) error
	DeleteCanvas(ctx context.Context, canvasId int, authUserId int) error
	ExecuteDynamicQuery(ctx context.Context, rawQuery string, authUserId int) ([]map[string]any, error)
}
