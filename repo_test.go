package model

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/yang-zzhong/go-querybuilder"
	. "testing"
	"time"
)

var con *sql.DB

type User struct {
	Id        string         `db:"id uuid[pk]"`
	Name      string         `db:"name varchar(32)"`
	Age       int            `db:"age int"`
	Level     int            `db:"level int"`
	Optional  sql.NullString `db:"optional varchar(256)[nil]"`
	CreatedAt time.Time      `db:"created_at datetime"`
}

func (u *User) TableName() string {
	return "users"
}

func init() {
	var err error
	con, err = sql.Open("mysql", "root:young159357789@/test_go?parseTime=true")
	if err != nil {
		fmt.Println(err)
	}
}

func TestRepo(t *T) {
	repo := NewRepo(&User{}, con, &MysqlModifier{})
	repo.Where("name", LIKE, "yang%")

	fmt.Println(repo.Fetch())
}
