package model

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	helpers "github.com/yang-zzhong/go-helpers"
	. "github.com/yang-zzhong/go-querybuilder"
	"reflect"
	"strings"
	. "testing"
	"time"
)

var con *sql.DB

type User struct {
	Id        string    `db:"id varchar(128) pk"`
	Name      string    `db:"name varchar(32) uk"`
	Age       int       `db:"age int"`
	Level     int       `db:"level int"`
	Optional  []string  `db:"optional varchar(256) nil"`
	CreatedAt time.Time `db:"created_at datetime"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) PK() string {
	return "id"
}

func (u *User) NewId() interface{} {
	return helpers.RandString(32)
}

func (u *User) DBValue(fieldName string, val interface{}) interface{} {
	switch fieldName {
	case "optional":
		return sql.NullString{strings.Join(val.([]string), ", "), true}
	default:
		return val
	}
	return val
}

func (u *User) Value(fieldName string, val interface{}) (result reflect.Value, catched bool) {
	if fieldName == "optional" {
		catched = true
		value, err := val.(sql.NullString).Value()
		if err != nil {
			fmt.Println(err)
		}
		if value != nil {
			result = reflect.ValueOf(strings.Split(value.(string), ", "))
			return
		}
		result = reflect.ValueOf([]string{})
		return
	}
	result = reflect.ValueOf(nil)
	catched = false
	return
}

func init() {
	var err error
	con, err = sql.Open("mysql", "root:xxx@/test_go?parseTime=true")
	if err != nil {
		fmt.Println(err)
	}
}

func TestRepo(t *T) {
	repo := NewRepo(&User{}, con, &MysqlModifier{})
	// fmt.Println(repo.CreateTable())
	// repo.Where("name", LIKE, "yang%")
	// repo.Create(user)
	// user := new(User)
	// user.Id = "y-zhong--"
	// user.Optional = []string{"hello", "world"}
	// user.Name = "y----yangzhong"
	// user.Age = 15
	// user.CreatedAt = time.Now()
	// if err := repo.Create(user); err != nil {
	// 	fmt.Println(err)
	// }

	fmt.Println(repo.Find("y-zhong--"))
	// repo = NewRepo(&User{}, con, &MysqlModifier{})
	// user := &User{"3", "y-zhong", 25, 2, sql.NullString{"", true}, time.Now()}
	// repo.Update(user)

	fmt.Println(repo.Find("3"))

	fmt.Println(repo.Fetch())
}
