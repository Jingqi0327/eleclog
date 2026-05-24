package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Querier
	UpdateRoomLastNotifiedAtTx(ctx context.Context, arg UpdateUserRoomLastNotifiedAtTxParams) (UserRoom, error)
	BindRoomToUserTx(ctx context.Context, arg BindRoomToUserTxParams) (BindRoomToUserTxResult, error)
}

type SQLStore struct {
	*Queries
	connPool *pgxpool.Pool
}

func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		Queries:  New(connPool),
		connPool: connPool,
	}
}
