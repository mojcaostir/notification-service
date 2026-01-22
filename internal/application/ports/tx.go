package ports

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Tx interface {
	pgx.Tx
}

type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error
}
