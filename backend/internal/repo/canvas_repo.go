package repo

import (
	"context"
	"fmt"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type canvasRepo struct {
	pool        DBTX
	prefixError string
}

func NewCanvasRepository(pool *pgxpool.Pool) model.CanvasRepository {
	return &canvasRepo{pool: pool, prefixError: "canvasRepo"}
}

func (r *canvasRepo) db(ctx context.Context) DBTX {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *canvasRepo) GetAll(ctx context.Context, active bool) ([]model.Canvas, error) {
	const fname = "GetAll"
	query := `
		SELECT 
			canvas_id, 
			canvas_name
		FROM canvas 
		WHERE ($1 = false) OR ($1 = true AND deleted_at IS NULL)
		`
	rows, err := r.db(ctx).Query(ctx, query, active)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	canvasList, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Canvas])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return canvasList, nil
}

func (r *canvasRepo) GetById(ctx context.Context, id int) (*model.Canvas, error) {
	const fname = "GetById"
	canvas := &model.Canvas{}
	query := "SELECT canvas_id, canvas_name, canvas_style FROM canvas WHERE canvas_id = $1"
	err := r.db(ctx).QueryRow(ctx, query, id).Scan(&canvas.CanvasId, &canvas.CanvasName, &canvas.CanvasStyle)

	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return canvas, nil
}

func (r *canvasRepo) GetByIds(ctx context.Context, ids []int) ([]model.Canvas, error) {
	const fname = "GetByIds"
	query := "SELECT canvas_id, canvas_name, canvas_style FROM canvas WHERE canvas_id = ANY($1)"
	rows, err := r.db(ctx).Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	canvasList, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Canvas])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return canvasList, nil
}

func (r *canvasRepo) GetCanvasByUserId(ctx context.Context, userId int, active bool) ([]model.Canvas, error) {
	const fname = "GetCanvasByUserId"
	query := `
		SELECT 
			c.canvas_id,
			c.canvas_name,
			c.canvas_style
		FROM "user" u
		JOIN canvas_role cr on cr.role_id = u.role_id
		JOIN canvas c on cr.canvas_id = c.canvas_id
		WHERE 
			u.user_id = $1 AND
			(($2 = false) OR ($2 = true AND c.deleted_at IS NULL))
	`
	rows, err := r.db(ctx).Query(ctx, query, userId, active)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	canvasResult, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Canvas])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return canvasResult, nil
}

func (r *canvasRepo) GetCanvasByRoleId(ctx context.Context, roleId int, active bool) ([]model.Canvas, error) {
	const fname = "GetCanvasByRoleId"
	query := `
		SELECT 
			c.canvas_id,
			c.canvas_name
		FROM canvas_role cr
		JOIN canvas c on cr.canvas_id = c.canvas_id
		WHERE 
			cr.role_id = $1 AND
			(($2 = false) OR ($2 = true AND c.deleted_at IS NULL))
	`
	rows, err := r.db(ctx).Query(ctx, query, roleId, active)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	canvasResult, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Canvas])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return canvasResult, nil
}

func (r *canvasRepo) GetAllCanvasRole(ctx context.Context) ([]model.CanvasRole, error) {
	const fname = "GetAllCanvasRole"
	query := `
		SELECT 
			cr.canvas_id,
			cr.role_id
		FROM canvas_role cr
	`
	rows, err := r.db(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	canvasResult, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.CanvasRole])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return canvasResult, nil
}

func (r *canvasRepo) GetCanvasRoleByRoleId(ctx context.Context, roleId int) ([]int, error) {
	const fname = "GetAllCanvasRole"
	query := `
		SELECT 
			canvas_id
		FROM canvas_role
		WHERE role_id = $1
	`
	rows, err := r.db(ctx).Query(ctx, query, roleId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	canvasIds, err := pgx.CollectRows(rows, pgx.RowTo[int])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return canvasIds, nil
}
func (r *canvasRepo) CreateCanvasRole(ctx context.Context, canvasroles []model.CanvasRole) error {
	const fname = "CreateCanvasRole"
	if len(canvasroles) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
        INSERT INTO canvas_role (
			role_id,
			canvas_id
        ) 
        VALUES ($1, $2) 
		`

	for _, canvasRole := range canvasroles {
		batch.Queue(
			query,
			canvasRole.RoleId,
			canvasRole.CanvasId,
		)
	}

	br := r.db(ctx).SendBatch(ctx, batch)
	defer br.Close()

	// br.Exec() returns the command tag (rows affected) and an error
	for range canvasroles {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
		}
	}

	return nil
}
func (r *canvasRepo) DeleteCanvasRole(ctx context.Context, canvasroles []model.CanvasRole) error {
	const fname = "DeleteCanvasRole"
	if len(canvasroles) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
        DELETE FROM canvas_role WHERE role_id = $1 AND canvas_id = $2	
		`

	for _, canvasRole := range canvasroles {
		batch.Queue(
			query,
			canvasRole.RoleId,
			canvasRole.CanvasId,
		)
	}

	br := r.db(ctx).SendBatch(ctx, batch)
	defer br.Close()

	for range canvasroles {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
		}
	}

	return nil
}

func (r *canvasRepo) Create(ctx context.Context, canvas *model.Canvas) error {
	const fname = "Create"
	query := `
		INSERT INTO canvas (canvas_name,canvas_style)
		VALUES ($1,$2)
		RETURNING canvas_id
	`
	err := r.db(ctx).QueryRow(ctx, query, canvas.CanvasName, canvas.CanvasStyle).Scan(&canvas.CanvasId)

	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return nil

}

func (r *canvasRepo) Update(ctx context.Context, canvas *model.Canvas) error {
	const fname = "Update"
	query := `
		UPDATE canvas
		SET 
			canvas_name = $1,
			canvas_style = $3
		WHERE canvas_id = $2		
	`
	result, err := r.db(ctx).Exec(ctx, query, canvas.CanvasName, canvas.CanvasId, canvas.CanvasStyle)

	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}

	return nil
}

func (r *canvasRepo) Delete(ctx context.Context, canvasId int) error {
	const fname = "Delete"
	query := `
		DELETE FROM canvas
		WHERE canvas_id = $1
	`
	result, err := r.db(ctx).Exec(ctx, query, canvasId)

	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}

	return nil
}

func (r *canvasRepo) CountValidate(ctx context.Context, canvas *model.Canvas) (int, error) {
	const fname = "CountValidate"
	var count int
	query := `
		SELECT count(*) 
		FROM canvas 
		WHERE 
			canvas_name = $1 AND 
			canvas_id != $2
	`
	err := r.db(ctx).QueryRow(ctx, query, canvas.CanvasName, canvas.CanvasId).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return count, nil
}

func (r *canvasRepo) ExecuteDynamicQuery(ctx context.Context, rawQuery string) ([]map[string]any, error) {
	const fname = "ExecuteDynamicQuery"

	// r.db(ctx) will automatically pull the Read-Only transaction from the context!
	rows, err := r.db(ctx).Query(ctx, rawQuery)
	if err != nil {
		return nil, fmt.Errorf("[%s] query execution failed: %w", fname, err)
	}
	defer rows.Close()

	// ⚡ Dynamically map the columns to Go types
	results, err := pgx.CollectRows(rows, pgx.RowToMap)
	if err != nil {
		return nil, fmt.Errorf("[%s] failed to parse dynamic rows: %w", fname, err)
	}

	if results == nil {
		results = []map[string]any{}
	}

	return results, nil
}
