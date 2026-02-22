package database

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type txKey struct{ db *DB }

func (db *DB) Transaction(ctx context.Context, fn func(context.Context) error) error {
	if _, ok := db.TxFromContext(ctx); ok {
		return fn(ctx)
	}

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txCtx := context.WithValue(ctx, txKey{db: db}, tx)
	if err := fn(txCtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (db *DB) TxFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{db: db}).(pgx.Tx)
	return tx, ok
}
