package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type DB struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, cfg Config) (*DB, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.connString())
	if err != nil {
		return nil, fmt.Errorf("database: parse config: %w", err)
	}

	if cfg.Pool.MaxConns > 0 {
		poolCfg.MaxConns = cfg.Pool.MaxConns
	}

	if cfg.Pool.MinConns > 0 {
		poolCfg.MinConns = cfg.Pool.MinConns
	}

	if cfg.Pool.MaxConnLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.Pool.MaxConnLifetime
	}

	if cfg.Pool.MaxConnIdleTime > 0 {
		poolCfg.MaxConnIdleTime = cfg.Pool.MaxConnIdleTime
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("database: create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database: ping: %w", err)
	}

	return &DB{pool: pool}, nil
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

func (db *DB) querier(ctx context.Context) Querier {
	if tx, ok := db.TxFromContext(ctx); ok {
		return tx
	}
	return db.pool
}
