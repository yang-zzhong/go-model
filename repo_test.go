package model

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/yang-zzhong/go-querybuilder"
	. "testing"
)

var con *sql.DB

type User struct {
	Id        string
	Name      string
	Age       int
	Level     int
	Optional  string
	CreatedAt []uint8 // time.Time
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) IdKey() interface{} {
	return "id"
}

func init() {
	var err error
	con, err = sql.Open("mysql", "root:young159357789@/test_go")
	if err != nil {
		fmt.Println(err)
	}
}

func TestRepo(t *T) {
	repo := NewRepo(&User{}, con, &MysqlModifier{})
	repo.Where("name", LIKE, "yang%")
	for _, item := range repo.Fetch() {
		fmt.Println(item)
	}
}
