package model

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type TransactionManager interface {
	Begin(ctx context.Context) (Transaction, error)
	// ⚡ NEW: Allows passing custom transaction options
	BeginTx(ctx context.Context, opts pgx.TxOptions) (Transaction, error)
}

type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Context() context.Context
}
