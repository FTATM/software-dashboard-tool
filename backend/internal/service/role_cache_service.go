package service

import (
	"sync"

	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type roleCache struct {
	store     sync.Map // map[roleId int]map[string][]string
	userRoles sync.Map // map[userId int]roleId int
}

// NewRoleCache creates a new instance of the in-memory permission cache
func NewRoleCache() model.RoleCache {
	return &roleCache{}
}

func (c *roleCache) Get(roleId int) (map[string][]string, bool) {
	if val, ok := c.store.Load(roleId); ok {
		return val.(map[string][]string), true
	}
	return nil, false
}

func (c *roleCache) Set(roleId int, permissions map[string][]string) {
	c.store.Store(roleId, permissions)
}

func (c *roleCache) Delete(roleId int) {
	c.store.Delete(roleId)
}

func (c *roleCache) GetUserRole(userId int) (int, bool) {
	if val, ok := c.userRoles.Load(userId); ok {
		return val.(int), true
	}
	return 0, false
}

func (c *roleCache) SetUserRole(userId int, roleId int) {
	c.userRoles.Store(userId, roleId)
}

func (c *roleCache) DeleteUserRole(userId int) {
	c.userRoles.Delete(userId)
}

func (c *roleCache) ClearUserRoles() {
	c.userRoles.Range(func(key, _ any) bool {
		c.userRoles.Delete(key)
		return true
	})
}
