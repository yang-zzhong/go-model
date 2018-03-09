package model

type FieldType interface {
	FieldType() string
	ConvertToFieldValue() string
}

// Base Type

type DBInt int64

func (i DBInt) FieldType() string {

}
