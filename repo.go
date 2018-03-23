package model

import (
	"database/sql"
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
	repo := &Repo{m, conn, NewModelMapper(m), NewBuilder(p)}
	repo.From(repo.model.(TableNamer).TableName())

	return repo
}

func (repo *Repo) Fetch() (result []interface{}, err error) {
	result = []interface{}{}
	rows, qerr := repo.conn.Query(repo.ForQuery(), repo.Params()...)
	if qerr != nil {
		err = qerr
		return
	}
	for rows.Next() {
		columns, cerr := rows.Columns()
		if cerr != nil {
			err = cerr
			return
		}
		rerr := rows.Scan(repo.mm.ValueReceivers(columns)...)
		if rerr != nil {
			err = rerr
			return
		}
		model := reflect.ValueOf(repo.mm.Model()).Elem().Interface()
		result = append(result, model)
	}

	return
}

func (repo *Repo) Update(data map[string]string) {
	repo.conn.Exec(repo.ForUpdate(data), repo.Params()...)
}

func (repo *Repo) Remove() {
	repo.conn.Exec(repo.ForRemove(), repo.Params()...)
}

func (repo *Repo) Create() {

}
