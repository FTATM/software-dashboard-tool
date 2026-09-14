package repo

import (
	"context"
	"fmt"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type roleRepo struct {
	pool        DBTX
	prefixError string
}

func NewRoleRepository(pool *pgxpool.Pool) model.RoleRepository {
	return &roleRepo{pool: pool, prefixError: "roleRepo"}
}

func (r *roleRepo) db(ctx context.Context) DBTX {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *roleRepo) CreateRole(ctx context.Context, role *model.Role) error {
	const fname = "CreateRole"
	query := `
		INSERT INTO role (role_name) VALUES ($1)	RETURNING role_id
	`
	err := r.db(ctx).QueryRow(ctx, query, role.RoleName).Scan(&role.RoleId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return nil
}

func (r *roleRepo) UpdateRole(ctx context.Context, role *model.Role) error {
	const fname = "UpdateRole"
	query := `
		UPDATE role SET role_name = $1 WHERE role_id = $2
	`
	result, err := r.db(ctx).Exec(ctx, query, role.RoleName, role.RoleId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}

	return nil
}

func (r *roleRepo) DeleteRolePermission(ctx context.Context, roleId int, menusIds []int, actionIds []int) error {
	const fname = "DeleteRolePermission"
	query := `
		DELETE FROM role_permission 
		WHERE role_id = $1 
		AND (menu_id, action_id) NOT IN (
			SELECT * FROM UNNEST($2::int[], $3::int[])
		)
	`
	_, err := r.db(ctx).Exec(ctx, query, roleId, menusIds, actionIds)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return nil
}

func (r *roleRepo) DeleteAllRolePermissions(ctx context.Context, roleId int) error {
	const fname = "DeleteAllRolePermissions"
	query := `DELETE FROM role_permission WHERE role_id = $1`
	_, err := r.db(ctx).Exec(ctx, query, roleId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return nil
}

func (r *roleRepo) CreateRolePermission(ctx context.Context, roleId int, menusIds []int, actionIds []int) error {
	const fname = "CreateRolePermission"
	query := `
		INSERT INTO role_permission (role_id, menu_id, action_id)
		SELECT $1, * FROM UNNEST($2::int[], $3::int[])
		ON CONFLICT (role_id, menu_id, action_id) DO NOTHING
	`
	_, err := r.db(ctx).Exec(ctx, query, roleId, menusIds, actionIds)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return nil
}

func (r *roleRepo) RolePermission(ctx context.Context, userId int, menuName string, actionName string) (bool, error) {
	const fname = "RolePermission"
	query := `
		SELECT EXISTS (
    		SELECT 1 
    		FROM "user" u
    		JOIN role_permission rp ON u.role_id = rp.role_id
    		JOIN menu m ON rp.menu_id = m.menu_id
    		JOIN action a ON rp.action_id = a.action_id
    		WHERE u.user_id = $1 
      		AND m.menu_name = $2 
      		AND a.action_name = $3
		);
	`
	var hasAccess bool
	err := r.db(ctx).QueryRow(ctx, query, userId, menuName, actionName).Scan(&hasAccess)
	if err != nil {
		return false, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return hasAccess, nil

}

func (r *roleRepo) GetAll(ctx context.Context) ([]model.Role, error) {
	const fname = "GetAll"
	query := `
		SELECT role_id, role_name FROM role
	`

	rows, err := r.db(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	roles, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Role])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return roles, nil

}

func (r *roleRepo) GetPermissionById(ctx context.Context, roleId int) (*model.RoleDetail, error) {
	const fname = "GetPermissionById"
	query := `
		SELECT 
			r.role_id,
			r.role_name,
			COALESCE(
				(
					SELECT json_agg(
						json_build_object(
							'roleId', rp.role_id,
							'menuId', rp.menu_id, 
							'actionId', rp.action_id
						)
					)
					FROM role_permission rp
					WHERE rp.role_id = r.role_id
				), 
				'[]'::json
			) AS permissions
		FROM role r
		WHERE r.role_id = $1;
	`

	rows, err := r.db(ctx).Query(ctx, query, roleId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	detail, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[model.RoleDetail])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return &detail, nil
}

func (r *roleRepo) GetMenuActionAvailable(ctx context.Context) ([]model.MainMenu, error) {
	const fname = "GetMenuActionAvailable"
	query := `
		SELECT 
			parent.menu_id,
			parent.menu_name,
	
			-- FETCH ACTIONS FOR FLAT MENUS (1D)
			COALESCE((
				SELECT json_agg(json_build_object('actionId', a.action_id, 'actionName', a.action_name))
				FROM menu_action ma
				JOIN action a ON ma.action_id = a.action_id
				WHERE ma.menu_id = parent.menu_id
			), '[]'::json) AS available_actions,

			-- FETCH SUBMENUS (2D)
			COALESCE(
				json_agg(
					json_build_object(
						'menuId', child.menu_id,
						'menuName', child.menu_name,
						'availableActions', COALESCE((
							SELECT json_agg(json_build_object('actionId', a.action_id, 'actionName', a.action_name))
							FROM menu_action ma
							JOIN action a ON ma.action_id = a.action_id
							WHERE ma.menu_id = child.menu_id
						), '[]'::json)
					)
				) FILTER (WHERE child.menu_id IS NOT NULL), '[]'::json
			) AS submenus

		FROM menu parent
		LEFT JOIN menu child ON child.parent_id = parent.menu_id
		WHERE parent.parent_id IS NULL 
		GROUP BY parent.menu_id, parent.menu_name
		ORDER BY parent.menu_id;
	`

	rows, err := r.db(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	// Updated to map directly into your new model.MainMenu slice
	mainMenus, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.MainMenu])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return mainMenus, nil
}

func (r *roleRepo) GetPermissionDescByRoleId(ctx context.Context, roleId int) ([]model.PermissionDesc, error) {
	const fname = "GetPermissionDescByRoleId"
	query := `
		SELECT 
    		m.menu_name, 
    		a.action_name
		FROM role_permission rp
		JOIN menu m ON rp.menu_id = m.menu_id
		JOIN action a ON rp.action_id = a.action_id
		WHERE rp.role_id = $1;
	`

	rows, err := r.db(ctx).Query(ctx, query, roleId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.PermissionDesc])

	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return results, nil
}

func (r *roleRepo) GetPermissionDescByUserId(ctx context.Context, userId int) ([]model.PermissionDesc, error) {
	const fname = "GetPermissionDescByUserId"
	query := `
		SELECT 
    		m.menu_name, 
    		a.action_name
		FROM "user" u
		JOIN role_permission rp ON u.role_id = rp.role_id
		JOIN menu m ON rp.menu_id = m.menu_id
		JOIN action a ON rp.action_id = a.action_id
		WHERE u.user_id = $1;
	`

	rows, err := r.db(ctx).Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.PermissionDesc])

	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return results, nil
}

func (r *roleRepo) CountUsersByRoleId(ctx context.Context, roleId int) (int, error) {
	const fname = "CountUsersByRoleId"
	// Using "user" table name as defined in your existing RolePermission query
	query := `
		SELECT COUNT(1) 
		FROM "user" 
		WHERE role_id = $1 AND deleted_at IS NULL
	`

	var count int
	err := r.db(ctx).QueryRow(ctx, query, roleId).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return count, nil
}

func (r *roleRepo) DeleteRole(ctx context.Context, roleId int) error {
	const fname = "DeleteRole"
	query := `DELETE FROM role WHERE role_id = $1`

	_, err := r.db(ctx).Exec(ctx, query, roleId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return nil
}
