package oracle

import (
	"context"

	sqlbuilder "github.com/huandu/go-sqlbuilder"
	_ "github.com/sijms/go-ora/v2"

	"github.com/kitti12911/lib-database/dbsql"
)

func New(ctx context.Context, cfg Config) (*dbsql.DB, error) {
	return dbsql.Open(ctx, "oracle", cfg.connString(), sqlbuilder.Oracle, cfg.Pool)
}
