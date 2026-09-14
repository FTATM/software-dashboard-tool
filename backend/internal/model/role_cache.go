package model

type RoleCache interface {
	Get(roleId int) (map[string][]string, bool)
	Set(roleId int, permissions map[string][]string)
	Delete(roleId int)
}
