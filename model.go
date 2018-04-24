package model

type Model interface {
	TableName() string // table name in db
	PK() string        // primary key
}

type ValueConverter interface {
	DBValue(name string) interface{}
}
