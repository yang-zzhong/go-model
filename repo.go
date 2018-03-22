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
	mm    *ModelMapper
	*Builder
}

var conn *sql.DB

func NewRepo(m interface{}, conn *sql.DB, p Modifier) *Repo {
	repo := &Repo{m, conn, NewMM(m), NewBuilder(p)}
	repo.From(repo.model.(TableNamer).TableName())

	return repo
}

func (repo *Repo) Fetch() []interface{} {
	rows, err := repo.conn.Query(repo.ForQuery(), repo.Params()...)
	if err != nil {
		fmt.Println(err)
	}
	result := []interface{}{}
	for rows.Next() {
		err = rows.Scan(repo.mm.FieldReceivers()...)
		result = append(result, repo.mm.Model())
	}

	return result
}

func (repo *Repo) pointers() []interface{} {
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
