package model

type Model interface {
	TableName() string  // table name in db
	PK() string         // primary key
	NewId() interface{} // NewId
}
