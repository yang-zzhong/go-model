package model

import (
	"database/sql"
	. "testing"
	"time"
)

type TestUser struct {
	Id        string    `db:"id varchar(128) pk"`
	Name      string    `db:"name varchar(32) uk"`
	Age       int       `db:"age int"`
	Level     int       `db:"level int nil"`
	Optional  string    `db:"optional varchar(256) nil"`
	CreatedAt time.Time `db:"created_at datetime"`
	UpdatedAt time.Time `db:"updated_at datetime nil"`
	*Base
}

func (u *TestUser) PK() string {
	return "id"
}

func (u *TestUser) TableName() string {
	return "users"
}

func TestCols(t *T) {
	mm := NewModelMapper(NewModel(new(TestUser)))
	_, err := mm.cols([]string{"id", "name", "age", "level", "optional", "created_at", "updated_at"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPack(t *T) {
	mm := NewModelMapper(NewModel(new(TestUser)))
	cols := []string{"id", "name", "age", "level", "optional", "created_at", "updated_at"}
	id := "1"
	name := "yang-zhong"
	age := 15
	level := sql.NullInt64{0, false}
	optional := sql.NullString{"hello", true}
	created_at := time.Now()
	updated_at := NullTime{time.Now(), false}
	res := []interface{}{&id, &name, &age, &level, &optional, &created_at, &updated_at}

	if _, _, err := mm.Pack(cols, res, new(TestUser).PK()); err != nil {
		t.Fatal(err)
	}
}
