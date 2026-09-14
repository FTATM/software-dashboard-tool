package model

type RolePermission struct {
	RoleId   int `json:"roleId,omitempty" db:"role_id"`
	MenuId   int `json:"menuId" db:"menu_id"`
	ActionId int `json:"actionId" db:"action_id"`
}

type PermissionDesc struct {
	MenuName   string `json:"menuName" db:"menu_name"`
	ActionName string `json:"actionName" db:"action_name"`
}
