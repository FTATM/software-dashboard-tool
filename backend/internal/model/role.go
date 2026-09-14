package model

import (
	"context"
	"time"
)

type Role struct {
	RoleId    int       `json:"roleId" db:"role_id"`
	RoleName  string    `json:"roleName" db:"role_name"`
	CreatedAt time.Time `json:"-" db:"created_at"`
}

type RoleDetail struct {
	RoleId          int              `json:"roleId" db:"role_id"`
	RoleName        string           `json:"roleName" db:"role_name"`
	RolePermissions []RolePermission `json:"rolePermissions" db:"permissions"`
}

type Access struct {
	UserId     int
	MenuName   string
	ActionName string
}

type UpsertRole struct {
	Role
	RolePermissions []RolePermission `json:"rolePermissions"`
}

type RoleRepository interface {
	CreateRole(ctx context.Context, role *Role) error
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRolePermission(ctx context.Context, roleId int, menusIds []int, actionIds []int) error
	CreateRolePermission(ctx context.Context, roleId int, menusIds []int, actionIds []int) error
	RolePermission(ctx context.Context, userId int, menuName string, actionName string) (bool, error)
	GetAll(ctx context.Context) ([]Role, error)
	GetPermissionById(ctx context.Context, roleId int) (*RoleDetail, error)
	GetMenuActionAvailable(ctx context.Context) ([]MainMenu, error)
	DeleteAllRolePermissions(ctx context.Context, roleId int) error
	GetPermissionDescByRoleId(ctx context.Context, roleId int) ([]PermissionDesc, error)
	GetPermissionDescByUserId(ctx context.Context, userId int) ([]PermissionDesc, error)
	CountUsersByRoleId(ctx context.Context, roleId int) (int, error)
	DeleteRole(ctx context.Context, roleId int) error
}

type RoleService interface {
	Access(ctx context.Context, acc *Access) (bool, error)
	UpsertRole(ctx context.Context, upsertRole *UpsertRole, authUserId int) error
	GetAll(ctx context.Context) ([]Role, error)
	GetMenuActionAvailable(ctx context.Context) ([]MainMenu, error)
	GetDetailById(ctx context.Context, roleId int) (*RoleDetail, error)
	DeleteRole(ctx context.Context, roleId int, authUserId int) error
}
