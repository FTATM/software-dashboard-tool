package model

type CanvasRole struct {
	RoleId   int `json:"roleId" db:"role_id"`
	CanvasId int `json:"canvasId" db:"canvas_id"`
}

type CanvasRoleDetail struct {
	RoleId    int   `json:"roleId"`
	CanvasIds []int `json:"canvasIds"`
}
