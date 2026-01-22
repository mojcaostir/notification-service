package db

import (
	"context"
	"fmt"

	"inbox-service/internal/application/ports"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManagerPG struct {
	pool *pgxpool.Pool
}

func NewTxManagerPG(pool *pgxpool.Pool) *TxManagerPG {
	return &TxManagerPG{pool: pool}
}

func (m *TxManagerPG) WithTx(ctx context.Context, fn func(ctx context.Context, tx ports.Tx) error) error {
	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
