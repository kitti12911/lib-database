package dbsql

import (
	"database/sql"
	"reflect"
	"strings"
	"sync"
)

type fieldInfo struct {
	index []int
	name  string
}

var fieldCache sync.Map

func cachedStructFields(t reflect.Type) []fieldInfo {
	if v, ok := fieldCache.Load(t); ok {
		return v.([]fieldInfo)
	}

	fields := parseStructFields(t)
	actual, _ := fieldCache.LoadOrStore(t, fields)
	return actual.([]fieldInfo)
}

func parseStructFields(t reflect.Type) []fieldInfo {
	var fields []fieldInfo
	walkFields(t, nil, &fields)
	return fields
}

func walkFields(t reflect.Type, index []int, fields *[]fieldInfo) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		idx := make([]int, len(index)+1)
		copy(idx, index)
		idx[len(index)] = i

		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			walkFields(f.Type, idx, fields)
			continue
		}

		tag := f.Tag.Get("db")
		if tag == "-" {
			continue
		}

		if tag == "" {
			tag = f.Name
		} else if comma := strings.IndexByte(tag, ','); comma != -1 {
			tag = tag[:comma]
		}

		*fields = append(*fields, fieldInfo{index: idx, name: strings.ToLower(tag)})
	}
}

type columnMapping struct {
	indices [][]int
}

func buildColumnMapping(columns []string, fields []fieldInfo) columnMapping {
	byName := make(map[string][]int, len(fields))
	for _, f := range fields {
		byName[f.name] = f.index
	}

	indices := make([][]int, len(columns))
	for i, col := range columns {
		indices[i] = byName[strings.ToLower(col)]
	}
	return columnMapping{indices: indices}
}

func collectOneRow[T any](rows *sql.Rows) (*T, error) {
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, nil
	}

	if reflect.TypeFor[T]().Kind() != reflect.Struct {
		var result T
		if err := rows.Scan(&result); err != nil {
			return nil, err
		}
		return &result, rows.Err()
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	fields := cachedStructFields(reflect.TypeFor[T]())
	cm := buildColumnMapping(columns, fields)

	result, err := scanRow[T](rows, cm)
	if err != nil {
		return nil, err
	}

	return result, rows.Err()
}

func collectRows[T any](rows *sql.Rows) ([]T, error) {
	defer rows.Close()

	if reflect.TypeFor[T]().Kind() != reflect.Struct {
		var results []T
		for rows.Next() {
			var v T
			if err := rows.Scan(&v); err != nil {
				return nil, err
			}
			results = append(results, v)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return results, nil
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	fields := cachedStructFields(reflect.TypeFor[T]())
	cm := buildColumnMapping(columns, fields)

	var results []T
	for rows.Next() {
		item, err := scanRow[T](rows, cm)
		if err != nil {
			return nil, err
		}
		results = append(results, *item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func scanRow[T any](rows *sql.Rows, cm columnMapping) (*T, error) {
	var result T
	v := reflect.ValueOf(&result).Elem()

	dests := make([]any, len(cm.indices))
	for i, idx := range cm.indices {
		if idx != nil {
			dests[i] = v.FieldByIndex(idx).Addr().Interface()
		} else {
			dests[i] = new(any)
		}
	}

	if err := rows.Scan(dests...); err != nil {
		return nil, err
	}

	return &result, nil
}
