package model

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	helpers "github.com/yang-zzhong/go-helpers"
	. "github.com/yang-zzhong/go-querybuilder"
	. "testing"
	"time"
)

var con *sql.DB

type User struct {
	Id        string         `db:"id varchar(128) pk"`
	Name      string         `db:"name varchar(32) uk"`
	Age       int            `db:"age int"`
	Level     int            `db:"level int"`
	Optional  sql.NullString `db:"optional varchar(256) nil"`
	CreatedAt time.Time      `db:"created_at datetime"`
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

func init() {
	var err error
	con, err = sql.Open("mysql", "root:young159357789@/test_go?parseTime=true")
	if err != nil {
		fmt.Println(err)
	}
}

func TestRepo(t *T) {
	repo := NewRepo(&User{}, con, &MysqlModifier{})
	// fmt.Println(repo.CreateTable())
	// repo.Where("name", LIKE, "yang%")
	// repo.Create(user)
	// user.Id = "y-zhong"
	// repo.Update(user)
	// fmt.Println(repo.Find("3"))
	// repo = NewRepo(&User{}, con, &MysqlModifier{})
	// user := &User{"3", "y-zhong", 25, 2, sql.NullString{"", true}, time.Now()}
	// repo.Update(user)

	fmt.Println(repo.Find("3"))

	fmt.Println(repo.Fetch())
}
