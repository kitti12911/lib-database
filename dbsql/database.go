package dbsql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sqlbuilder "github.com/huandu/go-sqlbuilder"
)

type PoolConfig struct {
	MaxOpenConns    int           `mapstructure:"max_open_conns"     validate:"omitempty,gte=1"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"     validate:"omitempty,gte=0"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"  validate:"omitempty,gt=0"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time" validate:"omitempty,gt=0"`
}

type Querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type DB struct {
	db     *sql.DB
	flavor sqlbuilder.Flavor
}

func Open(ctx context.Context, driverName, dataSourceName string, flavor sqlbuilder.Flavor, pool PoolConfig) (*DB, error) {
	sqlDB, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("database: open: %w", err)
	}

	if pool.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(pool.MaxOpenConns)
	}

	if pool.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(pool.MaxIdleConns)
	}

	if pool.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(pool.ConnMaxLifetime)
	}

	if pool.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(pool.ConnMaxIdleTime)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("database: ping: %w", err)
	}

	return &DB{db: sqlDB, flavor: flavor}, nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) SqlDB() *sql.DB {
	return db.db
}

func (db *DB) querier(ctx context.Context) Querier {
	if tx, ok := db.TxFromContext(ctx); ok {
		return tx
	}
	return db.db
}
