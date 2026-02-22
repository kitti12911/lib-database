package dbsql

import (
	"context"

	sqlbuilder "github.com/huandu/go-sqlbuilder"
)

type Builder interface {
	Build() (string, []any)
}

func FindOne[T any](ctx context.Context, db *DB, sql string, args ...any) (*T, error) {
	rows, err := db.querier(ctx).QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return collectOneRow[T](rows)
}

func FindOneB[T any](ctx context.Context, db *DB, b Builder) (*T, error) {
	sql, args := b.Build()
	return FindOne[T](ctx, db, sql, args...)
}

func FindAll[T any](ctx context.Context, db *DB, sql string, args ...any) ([]T, error) {
	rows, err := db.querier(ctx).QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	results, err := collectRows[T](rows)
	if err != nil {
		return nil, err
	}

	if results == nil {
		results = []T{}
	}
	return results, nil
}

func FindAllB[T any](ctx context.Context, db *DB, b Builder) ([]T, error) {
	sql, args := b.Build()
	return FindAll[T](ctx, db, sql, args...)
}

func FindAndCountB[T any](ctx context.Context, db *DB, b *sqlbuilder.SelectBuilder, limit, offset int) ([]T, int64, error) {
	q := db.querier(ctx)

	countSb := db.flavor.NewSelectBuilder()
	countSb.Select("COUNT(*)")
	countSb.From(countSb.BuilderAs(b, "__sub"))

	countSQL, countArgs := countSb.Build()

	var total int64
	if err := q.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []T{}, 0, nil
	}

	dataSb := b.Clone()
	dataSb.Limit(limit).Offset(offset)

	dataSQL, dataArgs := dataSb.Build()

	rows, err := q.QueryContext(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, 0, err
	}

	results, err := collectRows[T](rows)
	if err != nil {
		return nil, 0, err
	}

	if results == nil {
		results = []T{}
	}
	return results, total, nil
}

func Exec(ctx context.Context, db *DB, sql string, args ...any) (int64, error) {
	result, err := db.querier(ctx).ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func ExecB(ctx context.Context, db *DB, b Builder) (int64, error) {
	sql, args := b.Build()
	return Exec(ctx, db, sql, args...)
}
