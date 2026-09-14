package service

import (
	"sync"

	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type roleCache struct {
	store sync.Map
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
