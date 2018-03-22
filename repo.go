package model

import (
	// "time"
	"database/sql"
	"fmt"
	. "github.com/yang-zzhong/go-querybuilder"
	"reflect"
)

type Repo struct {
	model interface{}
	conn  *sql.DB
	*Builder
}

var conn *sql.DB

func NewRepo(m interface{}, conn *sql.DB, p Modifier) *Repo {
	repo := &Repo{m, conn, NewBuilder(p)}
	repo.From(repo.model.(Model).TableName())

	return repo
}

func (repo *Repo) Fetch() []interface{} {
	rows, err := repo.conn.Query(repo.ForQuery(), repo.Params()...)
	if err != nil {
		fmt.Println(err)
	}
	result := []interface{}{}
	for rows.Next() {
		err = rows.Scan(repo.pointers()...)
		result = append(result, repo.rowValue())
	}

	return result
}

func (repo *Repo) pointers() []interface{} {
	value := reflect.ValueOf(repo.model).Elem()
	length := value.NumField()
	pointers := make([]interface{}, length)
	for i := 0; i < length; i++ {
		pointers[i] = value.Field(i).Addr().Interface()
	}

	return pointers
}

func (repo *Repo) rowValue() interface{} {
	return reflect.ValueOf(repo.model).Elem().Interface()
}

func (repo *Repo) Update(data map[string]string) {
	repo.conn.Exec(repo.ForUpdate(data), repo.Params()...)
}

func (repo *Repo) Remove() {
	repo.conn.Exec(repo.ForRemove(), repo.Params()...)
}

func (repo *Repo) Create(m interface{}) {

}
