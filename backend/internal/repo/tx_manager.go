package repo

import (
	"context"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) model.TransactionManager {
	return &txManager{pool: pool}
}

// txWrapper implements model.Transaction
type txWrapper struct {
	tx  pgx.Tx
	ctx context.Context
}

func (t *txWrapper) Commit(ctx context.Context) error   { return t.tx.Commit(ctx) }
func (t *txWrapper) Rollback(ctx context.Context) error { return t.tx.Rollback(ctx) }
func (t *txWrapper) Context() context.Context           { return t.ctx }

// Begin starts the Postgres transaction and returns our wrapper
func (tm *txManager) Begin(ctx context.Context) (model.Transaction, error) {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	txCtx := injectTx(ctx, tx)

	return &txWrapper{
		tx:  tx,
		ctx: txCtx,
	}, nil
}

// ⚡ NEW: BeginTx starts a transaction with specific options (like Read-Only)
func (tm *txManager) BeginTx(ctx context.Context, opts pgx.TxOptions) (model.Transaction, error) {
	tx, err := tm.pool.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	txCtx := injectTx(ctx, tx)

	return &txWrapper{
		tx:  tx,
		ctx: txCtx,
	}, nil
}
