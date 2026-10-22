package service

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type roleService struct {
	txManager    model.TransactionManager
	roleRepo     model.RoleRepository
	prefixError  string
	auditLogRepo model.AuditLogRepository
	roleCache    model.RoleCache
}

func NewRoleService(txManager model.TransactionManager, roleRepo model.RoleRepository, auditLogRepo model.AuditLogRepository, roleCache model.RoleCache) model.RoleService {
	return &roleService{
		txManager:    txManager,
		roleRepo:     roleRepo,
		prefixError:  "roleService",
		auditLogRepo: auditLogRepo,
		roleCache:    roleCache,
	}
}

func (s *roleService) Access(ctx context.Context, acc *model.Access) (bool, error) {
	const fname = "Access"

	// ⚡ Fast in-memory check when both user->role and role->permissions are cached
	if s.roleCache != nil && acc != nil && acc.UserId > 0 {
		if roleId, ok := s.roleCache.GetUserRole(acc.UserId); ok {
			if permMap, ok := s.roleCache.Get(roleId); ok {
				return slices.Contains(permMap[acc.MenuName], acc.ActionName), nil
			}
		}
	}

	hasAccess, err := s.roleRepo.RolePermission(ctx, acc.UserId, acc.MenuName, acc.ActionName)
	if err != nil {
		return false, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return hasAccess, nil
}

func (s *roleService) UpsertRole(ctx context.Context, upsertRole *model.UpsertRole, authUserId int) error {
	const fname = "UpsertRole"
	var menuIds []int
	var actionIds []int
	var err error
	for _, p := range upsertRole.RolePermissions {
		menuIds = append(menuIds, p.MenuId)
		actionIds = append(actionIds, p.ActionId)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	if upsertRole.RoleId == 0 {
		err = s.roleRepo.CreateRole(tx.Context(), &upsertRole.Role)
	} else {
		err = s.roleRepo.UpdateRole(tx.Context(), &upsertRole.Role)
	}

	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if len(menuIds) == 0 {
		err = s.roleRepo.DeleteAllRolePermissions(tx.Context(), upsertRole.RoleId)
		if err != nil {
			return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}
	} else {
		err = s.roleRepo.DeleteRolePermission(tx.Context(), upsertRole.RoleId, menuIds, actionIds)
		if err != nil {
			return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}

		err = s.roleRepo.CreateRolePermission(tx.Context(), upsertRole.RoleId, menuIds, actionIds)
		if err != nil {
			return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}
	}

	newData, err := model.StructToDynamicJSON(upsertRole)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	audit := model.AuditLog{
		EntityType: "role",
		EntityId:   strconv.Itoa(upsertRole.RoleId),
		MenuType:   "Role",
		Action:     model.UpdateAction,
		ChangedBy:  authUserId,
		OldData:    nil,
		NewData:    newData,
	}

	if err = s.auditLogRepo.Create(tx.Context(), []model.AuditLog{audit}); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if s.roleCache != nil {
		s.roleCache.Delete(upsertRole.RoleId)
	}

	return nil
}

func (s *roleService) GetAll(ctx context.Context) ([]model.Role, error) {
	const fname = "GetAll"
	roles, err := s.roleRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return roles, nil

}

func (s *roleService) GetMenuActionAvailable(ctx context.Context) ([]model.MainMenu, error) {
	const fname = "GetMenuActionAvailable"
	mainMenus, err := s.roleRepo.GetMenuActionAvailable(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return mainMenus, nil
}

func (s *roleService) GetDetailById(ctx context.Context, roleId int) (*model.RoleDetail, error) {
	const fname = "GetDetailById"
	detail, err := s.roleRepo.GetPermissionById(ctx, roleId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return detail, nil
}

func (s *roleService) DeleteRole(ctx context.Context, roleId int, authUserId int) error {
	const fname = "DeleteRole"

	userCount, err := s.roleRepo.CountUsersByRoleId(ctx, roleId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if userCount > 0 {
		return fmt.Errorf("Cannot delete this role because %d user(s) are currently assigned to it. %w", userCount, model.ErrInUsed)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	defer tx.Rollback(ctx)

	err = s.roleRepo.DeleteAllRolePermissions(tx.Context(), roleId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	err = s.roleRepo.DeleteRole(tx.Context(), roleId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	audit := model.AuditLog{
		EntityType: "role",
		EntityId:   strconv.Itoa(roleId),
		MenuType:   "Role",
		Action:     model.DeleteAction,
		ChangedBy:  authUserId,
		OldData:    nil,
		NewData:    nil,
	}

	if err = s.auditLogRepo.Create(tx.Context(), []model.AuditLog{audit}); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if s.roleCache != nil {
		s.roleCache.Delete(roleId)
		s.roleCache.ClearUserRoles()
	}

	return nil
}
