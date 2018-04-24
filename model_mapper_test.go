package model

import (
	"database/sql"
	. "testing"
	"time"
)

type TestUser struct {
	Id        string         `db:"id varchar(128) pk"`
	Name      string         `db:"name varchar(32) uk"`
	Age       int            `db:"age int"`
	Level     int            `db:"level int"`
	Optional  sql.NullString `db:"optional varchar(256) nil"`
	CreatedAt time.Time      `db:"created_at datetime"`
}

var now time.Time
var user *TestUser
var dbValues map[string]interface{}
var values map[string]interface{}
var mapper *ModelMapper

func init() {
	now = time.Now()
	user = new(TestUser)
	user.Id = "123456"
	user.Name = "test-name"
	user.Age = 25
	user.CreatedAt = now
	dbValues = map[string]interface{}{
		"id":         "123456",
		"name":       "test-name",
		"age":        25,
		"level":      0,
		"optional":   sql.NullString{"", false},
		"created_at": now,
	}
	values = map[string]interface{}{
		"Id":        "123456",
		"Name":      "test-name",
		"Age":       25,
		"Level":     0,
		"Optional":  sql.NullString{"", false},
		"CreatedAt": now,
	}
	mapper = NewModelMapper(new(TestUser))
}

func TestExtract(t *T) {
	result, err := mapper.Extract(user)
	if err != nil {
		panic(err)
	}
	for fieldName, value := range result {
		if value != dbValues[fieldName] {
			t.Fatalf(
				"Extract: %s's value error, error value is %v, should be %v",
				fieldName,
				value,
				dbValues[fieldName],
			)
		}
	}
}

func TestDBFieldValue(t *T) {
	for fieldName, value := range dbValues {
		dbValue, err := mapper.DbFieldValue(user, fieldName)
		if err != nil {
			panic(err)
		}
		if dbValue != value {
			t.Fatalf(
				"DbFieldValue: %s's value error, error value is %v, should be %v",
				fieldName,
				dbValue,
				value,
			)
		}
	}
}

func TestFieldValue(t *T) {
	for field, value := range values {
		tValue, err := mapper.FieldValue(user, field)
		if err != nil {
			panic(err)
		}
		if tValue != value {
			t.Fatalf(
				"FieldValue: %s's value error, error value is %v, should be %v",
				field,
				tValue,
				value,
			)
		}
	}
}
