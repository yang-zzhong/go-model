package model

type Model interface {
	TableName() string
	IdKey() interface{}
}

type BaseModel struct {}
