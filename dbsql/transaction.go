package dbsql

import (
	"context"
	"database/sql"
)

type txKey struct{ db *DB }

func (db *DB) Transaction(ctx context.Context, fn func(context.Context) error) error {
	if _, ok := db.TxFromContext(ctx); ok {
		return fn(ctx)
	}

	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txCtx := context.WithValue(ctx, txKey{db: db}, tx)
	if err := fn(txCtx); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *DB) TxFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey{db: db}).(*sql.Tx)
	return tx, ok
}
