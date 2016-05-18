package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"time"

	"gopkg.in/pg.v4/types"
)

var timeType = reflect.TypeOf((*time.Time)(nil)).Elem()

type useQueryOne interface {
	useQueryOne() bool
}

type Model interface {
	Collection
	ColumnScanner
}

func NewModel(values ...interface{}) (Model, error) {
	if len(values) > 1 {
		return Scan(values...), nil
	}

	v0 := values[0]
	switch v0 := v0.(type) {
	case Model:
		return v0, nil
	case sql.Scanner:
		return Scan(v0), nil
	}

	v := reflect.ValueOf(v0)
	if !v.IsValid() {
		return nil, errors.New("pg: Model(nil)")
	}
	if v.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("pg: Model(nonsettable %T)", v0)
	}
	v = v.Elem()

	switch v.Kind() {
	case reflect.Struct:
		return newStructTableModel(v)
	case reflect.Slice:
		elType := indirectType(v.Type().Elem())
		if elType.Kind() == reflect.Struct && elType != timeType {
			return &sliceTableModel{
				structTableModel: structTableModel{
					table: Tables.Get(elType),
					root:  v,
				},
				slice: v,
			}, nil
		} else {
			return &sliceModel{
				slice: v,
				scan:  types.Scanner(elType),
			}, nil
		}
	}

	return Scan(v0), nil
}
