package model

import (
	"database/sql"
	"errors"
	. "github.com/yang-zzhong/go-querybuilder"
	"reflect"
)

type onModify func(model interface{})

type conn struct {
	db *sql.DB
	m  Modifier
}

var (
	c      conn
	inited bool
)

const (
	t_one = iota
	t_many
)

type with struct {
	name string
	m    interface{}
	n    Nexus
	t    int
}

func Config(db *sql.DB, m Modifier) {
	c.db = db
	c.m = m
	inited = true
}

type Repo struct {
	model    interface{}
	conn     *sql.DB
	mm       *ModelMapper
	modifier Modifier
	onCreate onModify
	onUpdate onModify
	withs    []with
	tx       *sql.Tx
	*Builder
}

func NewCustomRepo(m interface{}, conn *sql.DB, p Modifier) *Repo {
	repo := new(Repo)
	repo.model = m
	repo.mm = NewModelMapper(m)
	repo.conn = conn
	repo.modifier = p
	repo.onCreate = func(model interface{}) {}
	repo.onUpdate = func(model interface{}) {}
	repo.Builder = NewBuilder(p)
	repo.withs = []with{}
	repo.From(repo.model.(Model).TableName())

	return repo
}

func NewRepo(m interface{}) (repo *Repo, err error) {
	if !inited {
		err = errors.New("not config the db and modifier yet")
		return
	}
	repo = NewCustomRepo(m, c.db, c.m)
	return
}

func (repo *Repo) Clean() {
	repo.Builder.Init()
}

func (repo *Repo) Fetch() (models map[interface{}]interface{}, err error) {
	var (
		rows    *sql.Rows
		columns []string
		cols    []interface{}
	)
	models = make(map[interface{}]interface{})
	if rows, err = repo.query(); err != nil {
		return
	}
	defer rows.Close()
	colget := false
	for rows.Next() {
		if !colget {
			if columns, err = rows.Columns(); err != nil {
				return
			}
			if cols, err = repo.mm.cols(columns); err != nil {
				return
			}
			colget = true
		}
		if err = rows.Scan(cols...); err != nil {
			return
		}
		var m, id interface{}
		m, id, err = repo.mm.Pack(columns, cols)
		if err != nil {
			return
		}
		models[id] = m
	}
	nexusValues := repo.nexusValues(models)
	for id, _ := range models {
		repo.bindNexus(models[id], nexusValues)
	}

	return
}

func (repo *Repo) One() (interface{}, error) {
	var rows map[interface{}]interface{}
	var err error
	if rows, err = repo.Fetch(); err != nil {
		return nil, err
	}
	for _, row := range rows {
		return row, nil
	}

	return nil, nil
}

func (repo *Repo) Find(id interface{}) (interface{}, error) {
	var rows map[interface{}]interface{}
	var err error
	r := NewCustomRepo(repo.model, repo.conn, repo.modifier)
	r.Where(repo.model.(Model).PK(), id).Limit(1)
	if rows, err = r.Fetch(); err != nil {
		return nil, err
	}
	for _, row := range rows {
		return row, nil
	}

	return nil, nil
}

func (repo *Repo) Update(models interface{}) error {
	var row map[string]interface{}
	var err error
	val := reflect.ValueOf(models)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	switch val.Kind() {
	case reflect.Struct:
		row = repo.mm.Extract(models)
		_, err = repo.exec(repo.ForUpdate(row))
		return err
	case reflect.Slice:
		slice := models.([]interface{})
		for _, m := range slice {
			row = repo.mm.Extract(m)
			_, err = repo.exec(repo.ForUpdate(row))
			return err
		}
	case reflect.Map:
		maps := models.([]interface{})
		for _, m := range maps {
			row = repo.mm.Extract(m)
			_, err = repo.exec(repo.ForUpdate(row))
			return err
		}
	}
	return nil
}

func (repo *Repo) Create(models interface{}) error {
	var err error
	var data []map[string]interface{}
	val := reflect.ValueOf(models)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	switch val.Kind() {
	case reflect.Struct:
		data = append(data, repo.mm.Extract(val.Interface()))
	case reflect.Slice:
		slice := val.Interface().([]interface{})
		for _, m := range slice {
			data = append(data, repo.mm.Extract(m))
		}
	case reflect.Map:
		maps := val.Interface().(map[interface{}]interface{})
		for _, m := range maps {
			data = append(data, repo.mm.Extract(m))
		}
	}
	_, err = repo.exec(repo.ForInsert(data))

	return err
}

func (repo *Repo) exec(sql string) (sql.Result, error) {
	if repo.tx != nil {
		return repo.tx.Exec(sql, repo.Params()...)
	}
	return repo.conn.Exec(sql, repo.Params()...)
}

func (repo *Repo) query() (*sql.Rows, error) {
	return repo.conn.Query(repo.ForQuery(), repo.Params()...)
}
