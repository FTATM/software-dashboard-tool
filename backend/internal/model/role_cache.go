package model

type RoleCache interface {
	Get(roleId int) (map[string][]string, bool)
	Set(roleId int, permissions map[string][]string)
	Delete(roleId int)
	GetUserRole(userId int) (int, bool)
	SetUserRole(userId int, roleId int)
	DeleteUserRole(userId int)
	ClearUserRoles()
}
