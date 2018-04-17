package model

type Model interface {
	TableName() string
	IdKey() string
}
