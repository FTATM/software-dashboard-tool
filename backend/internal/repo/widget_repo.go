package repo

import (
	"context"
	"fmt"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type widgetRepo struct {
	pool        DBTX
	prefixError string
}

func NewWidgetRepository(pool *pgxpool.Pool) model.WidgetRepository {
	return &widgetRepo{pool: pool, prefixError: "widgetRepo"}
}

func (r *widgetRepo) db(ctx context.Context) DBTX {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *widgetRepo) GetById(ctx context.Context, id int) (*model.Widget, error) {
	const fname = "GetById"
	widget := &model.Widget{}
	query := "SELECT widget_id, widget_type_id, layout_data, widget_label FROM widget WHERE widget_id = $1"
	err := r.db(ctx).QueryRow(ctx, query, id).Scan(&widget.WidgetId, &widget.WidgetTypeId, &widget.LayoutData)

	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return widget, nil
}

func (r *widgetRepo) Create(ctx context.Context, widgets []model.Widget) error {
	const fname = "Create"
	if len(widgets) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
        INSERT INTO widget (
            widget_type_id,
			canvas_id,
			device_id_s,
			widget_label,
    		layout_data,
			widget_style,
			custom_chart_data,
			device_group_id
        ) 
        VALUES ($1, $2, $3, $4, $5, $6, $7,$8) 
        RETURNING widget_id`

	// 1. Queue all the queries
	for _, widget := range widgets {
		batch.Queue(
			query,
			widget.WidgetTypeId,
			widget.CanvasId,
			widget.DeviceIds,
			widget.WidgetLabel,
			widget.LayoutData,
			widget.WidgetStyle,
			widget.CustomChartData,
			widget.DeviceGroupId,
		)
	}

	// 2. Send the batch in one network trip
	br := r.db(ctx).SendBatch(ctx, batch)
	defer br.Close() // Ensure the batch is closed

	// 3. Read the returned IDs in the exact same order
	for i := range widgets {
		err := br.QueryRow().Scan(&widgets[i].WidgetId)
		if err != nil {
			return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
		}
	}

	return nil
}

func (r *widgetRepo) Update(ctx context.Context, widgets []model.Widget) error {
	const fname = "Update"
	if len(widgets) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
        UPDATE widget 
        SET 
            widget_type_id = $1,
            canvas_id = $2,
            device_id_s = $3,
            widget_label = $4,
            layout_data = $5,
            widget_style = $6,
            custom_chart_data = $7,
			device_group_id = $9
        WHERE widget_id = $8`

	// 1. Queue all the individual update statements
	for _, w := range widgets {
		batch.Queue(query,
			w.WidgetTypeId,
			w.CanvasId,
			w.DeviceIds,
			w.WidgetLabel,
			w.LayoutData,
			w.WidgetStyle,
			w.CustomChartData,
			w.WidgetId,
			w.DeviceGroupId,
		)
	}

	// 2. Send to the database
	br := r.db(ctx).SendBatch(ctx, batch)
	defer br.Close()

	// 3. Verify that every update actually affected a row
	var totalRowsAffected int64
	for range widgets {
		tag, err := br.Exec() // Exec is used instead of QueryRow for updates
		if err != nil {
			return fmt.Errorf("[%s]>[%s] execute error: %w", r.prefixError, fname, err)
		}
		totalRowsAffected += tag.RowsAffected()
	}

	if totalRowsAffected != int64(len(widgets)) {
		return fmt.Errorf("[%s]>[%s]: update row affected not match (expected %d, got %d)",
			r.prefixError, fname, len(widgets), totalRowsAffected)
	}

	return nil
}

func (r *widgetRepo) Delete(ctx context.Context, ids []int) error {
	const fname = "Delete"
	if len(ids) == 0 {
		return nil
	}

	query := `DELETE FROM widget WHERE widget_id = ANY($1::INT[])`

	result, err := r.db(ctx).Exec(ctx, query, ids)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	if result.RowsAffected() != int64(len(ids)) {
		return fmt.Errorf("[%s]>[%s]: attempted to delete %d widgets, but only found and deleted %d", r.prefixError, fname, len(ids), result.RowsAffected())
	}

	return nil
}

func (r *widgetRepo) GetWidgetByCanvasId(ctx context.Context, canvasId []int) ([]model.Widget, error) {
	const fname = "GetWidgetByCanvasId"
	query := `
	SELECT 
		widget_id,
		widget_type_id,
		widget_label,
		canvas_id,
		device_id_s,
		layout_data,
		widget_style,
		custom_chart_data,
		device_group_id
	FROM widget 
	WHERE canvas_id = ANY($1)
	`
	rows, err := r.db(ctx).Query(ctx, query, canvasId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	widgets, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.Widget])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return widgets, nil
}

func (r *widgetRepo) GetActiveImageFilenames(ctx context.Context) ([]string, error) {
	// Extracts the imageUrl string from the JSONB field where it exists and is not empty
	query := `
		SELECT custom_chart_data->>'imageUrl'
		FROM widget
		WHERE custom_chart_data->>'imageUrl' IS NOT NULL 
		  AND custom_chart_data->>'imageUrl' != ''
	`
	rows, err := r.db(ctx).Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var filenames []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			continue
		}
		filenames = append(filenames, url)
	}
	return filenames, nil
}
