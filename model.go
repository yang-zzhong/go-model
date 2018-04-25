package model

import (
	"reflect"
)

type Model interface {
	TableName() string // table name in db
	PK() string        // primary key
}

type ValueConverter interface {
	DBValue(fieldName string, value interface{}) interface{}
	Value(fieldName string, value interface{}) (reflect.Value, bool)
}
