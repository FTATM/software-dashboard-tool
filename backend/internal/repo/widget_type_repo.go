package repo

import (
	"context"
	"fmt"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type widgetTypeRepo struct {
	pool        DBTX
	prefixError string
}

func NewWidgetTypeRepository(pool *pgxpool.Pool) model.WidgetTypeRepository {
	return &widgetTypeRepo{pool: pool, prefixError: "widgetTypeRepo"}
}

func (r *widgetTypeRepo) db(ctx context.Context) DBTX {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *widgetTypeRepo) GetById(ctx context.Context, id int) (*model.WidgetType, error) {
	const fname = "GetById"
	widget := &model.WidgetType{}
	query := "SELECT widget_type_id, widget_type_name FROM widget_type WHERE widget_type_id = $1"
	err := r.db(ctx).QueryRow(ctx, query, id).Scan(&widget.WidgetTypeId, &widget.WidgetTypeName)

	if err != nil {
		// if errors.Is(err, pgx.ErrNoRows) {
		// 	return nil,
		// }
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return widget, nil
}

func (r *widgetTypeRepo) GetAll(ctx context.Context) ([]model.WidgetType, error) {
	const fname = "GetAll"
	query := "SELECT widget_type_id, widget_type_name FROM widget_type"
	rows, err := r.db(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	widgetTypes, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.WidgetType])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return widgetTypes, nil
}
