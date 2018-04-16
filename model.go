package model

type TableNamer interface {
	TableName() string
	IdKey() string
}
