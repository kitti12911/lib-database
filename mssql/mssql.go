package mssql

import (
	"context"

	sqlbuilder "github.com/huandu/go-sqlbuilder"
	_ "github.com/microsoft/go-mssqldb"

	"github.com/kitti12911/lib-database/dbsql"
)

func New(ctx context.Context, cfg Config) (*dbsql.DB, error) {
	return dbsql.Open(ctx, "sqlserver", cfg.connString(), sqlbuilder.SQLServer, cfg.Pool)
}
