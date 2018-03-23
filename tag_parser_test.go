package model

import (
	"fmt"
	. "testing"
)

func TestDB(t *T) {
	tp := new(TagParser)
	db := "id varchar(10)[pk]"
	result, _ := tp.ParseDB(db)
	if result.FieldName != "id" {
		t.Error("id fail")
	}
	if result.FieldType != "varchar(10)" {
		t.Error("type fail")
	}
	if !result.IsPk {
		t.Error("pk fail")
	}
	db = "id[pk]"
	_, err := tp.ParseDB(db)
	if err == nil {
		t.Error("error fail")
	}
}

func TestUK(t *T) {
	tp := new(TagParser)
	uk := "hello,world"
	fmt.Println(tp.ParseUk(uk))
}

func TestFk(t *T) {
	tp := new(TagParser)
	fk := "hello,world"
	fmt.Println(tp.ParseUk(fk))
}
