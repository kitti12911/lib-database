package database

import (
	"context"
	"errors"
	"reflect"

	"github.com/jackc/pgx/v5"

	sqlbuilder "github.com/huandu/go-sqlbuilder"
)

type Builder interface {
	Build() (string, []any)
}

func isStruct[T any]() bool {
	return reflect.TypeFor[T]().Kind() == reflect.Struct
}

func FindOne[T any](ctx context.Context, db *DB, sql string, args ...any) (*T, error) {
	rows, err := db.querier(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	var collector pgx.RowToFunc[*T]
	if isStruct[T]() {
		collector = pgx.RowToAddrOfStructByName[T]
	} else {
		collector = pgx.RowToAddrOf[T]
	}

	result, err := pgx.CollectOneRow(rows, collector)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return result, err
}

func FindOneB[T any](ctx context.Context, db *DB, b Builder) (*T, error) {
	sql, args := b.Build()
	return FindOne[T](ctx, db, sql, args...)
}

func FindAll[T any](ctx context.Context, db *DB, sql string, args ...any) ([]T, error) {
	rows, err := db.querier(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	var collector pgx.RowToFunc[T]
	if isStruct[T]() {
		collector = pgx.RowToStructByName[T]
	} else {
		collector = pgx.RowTo[T]
	}

	results, err := pgx.CollectRows(rows, collector)
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

	countSb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	countSb.Select("COUNT(*)")
	countSb.From(countSb.BuilderAs(b, "__sub"))

	countSQL, countArgs := countSb.Build()

	var total int64
	if err := q.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []T{}, 0, nil
	}

	dataSb := b.Clone()
	dataSb.Limit(limit).Offset(offset)

	dataSQL, dataArgs := dataSb.Build()

	rows, err := q.Query(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, 0, err
	}

	var collector pgx.RowToFunc[T]
	if isStruct[T]() {
		collector = pgx.RowToStructByName[T]
	} else {
		collector = pgx.RowTo[T]
	}

	results, err := pgx.CollectRows(rows, collector)
	if err != nil {
		return nil, 0, err
	}

	if results == nil {
		results = []T{}
	}
	return results, total, nil
}

func Exec(ctx context.Context, db *DB, sql string, args ...any) (int64, error) {
	tag, err := db.querier(ctx).Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func ExecB(ctx context.Context, db *DB, b Builder) (int64, error) {
	sql, args := b.Build()
	return Exec(ctx, db, sql, args...)
}
