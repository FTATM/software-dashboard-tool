package repo

import (
	"context"
	"fmt"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type auditLogRepo struct {
	pool        DBTX
	prefixError string
}

func NewAuditLogRepository(pool *pgxpool.Pool) model.AuditLogRepository {
	return &auditLogRepo{pool: pool, prefixError: "auditLogRepo"}
}

func (r *auditLogRepo) db(ctx context.Context) DBTX {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *auditLogRepo) Create(ctx context.Context, logs []model.AuditLog) error {
	const fname = "Create"
	if len(logs) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
        INSERT INTO audit_log (
			entity_type,
			entity_id,
			action,
			changed_by,
			old_data,
			new_data,
			menu_type
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	for _, log := range logs {
		batch.Queue(
			query,
			log.EntityType,
			log.EntityId,
			log.Action,
			log.ChangedBy,
			log.OldData,
			log.NewData,
			log.MenuType,
		)
	}

	br := r.db(ctx).SendBatch(ctx, batch)
	defer br.Close() // Ensure the batch is closed

	for i := range logs {
		err := br.QueryRow().Scan(&logs[i].Id)
		if err != nil {
			return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
		}
	}

	return nil
}
