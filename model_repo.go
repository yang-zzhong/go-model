package model

import (
	"database/sql"
	"errors"
	. "github.com/yang-zzhong/go-querybuilder"
	"reflect"
)

type rowshandler func(*sql.Rows, []string)

// oncreate and onupdate callback type
type modify func(model interface{})

// connection struct
type conn struct {
	db *sql.DB
	m  Modifier
}

var (
	c      conn // Config set the conn as default conn
	inited bool // if inited
)

const (
	t_one  = 1 // relationship is a has one relationship
	t_many = 2 // relationship is a has many relationship
)

// a relationship with repo
type with struct {
	name string      // relationship name
	m    interface{} // relationship target
	n    Nexus       // relationship nexus
	t    int         // relationship type t_one|t_many
}

// config the default connected db
func Config(db *sql.DB, m Modifier) {
	c.db = db
	c.m = m
	inited = true
}

// repo
type Repo struct {
	model    interface{} // repo row model
	conn     *sql.DB     // conn
	modifier Modifier    // sql modifier
	oncreate modify      // on create callback
	onupdate modify      // on update callback
	withs    []with      // maintain fetch model relationship
	tx       *sql.Tx     // tx
	*Builder
}

// new custom repo
func NewCustomRepo(m interface{}, conn *sql.DB, p Modifier) *Repo {
	repo := new(Repo)
	repo.model = m
	repo.conn = conn
	repo.modifier = p
	repo.oncreate = func(model interface{}) {}
	repo.onupdate = func(model interface{}) {}
	repo.Builder = NewBuilder(p)
	repo.withs = []with{}
	repo.From(repo.model.(Model).TableName())

	return repo
}

// new default repo
func NewRepo(m interface{}) (repo *Repo, err error) {
	if !inited {
		err = errors.New("not config the db and modifier yet")
		return
	}
	repo = NewCustomRepo(m, c.db, c.m)
	return
}

// set on create callback
func (repo *Repo) OnCreate(c modify) *Repo {
	repo.oncreate = c
	return repo
}

// set on update callback
func (repo *Repo) OnUpdate(c modify) *Repo {
	repo.onupdate = c
	return repo
}

// clean builder
func (repo *Repo) Clean() {
	repo.Builder.Init()
}

// count
func (repo *Repo) Count() (int, error) {
	rows, err := repo.conn.Query(repo.ForCount(), repo.Params()...)
	if err != nil {
		return 0, err
	}
	var count int
	for rows.Next() {
		rows.Scan(&count)
		break
	}
	return count, nil
}

func (repo *Repo) Query(handle rowshandler) error {
	var rows *sql.Rows
	var cols []string
	var err error
	if rows, err = repo.query(); err != nil {
		return err
	}
	defer rows.Close()
	if cols, err = rows.Columns(); err != nil {
		return err
	}
	for rows.Next() {
		handle(rows, cols)
	}

	return nil
}

func (repo *Repo) Fetch() (models map[interface{}]interface{}, err error) {
	return repo.FetchKey(repo.model.(Model).PK())
}

func (repo *Repo) FetchKey(col string) (models map[interface{}]interface{}, err error) {
	var cols []interface{}
	models = make(map[interface{}]interface{})
	colget := false
	err = repo.Query(func(rows *sql.Rows, columns []string) {
		if !colget {
			if cols, err = repo.model.(Mapable).Mapper().cols(columns); err != nil {
				return
			}
			colget = true
		}
		if err = rows.Scan(cols...); err != nil {
			return
		}
		var m, id interface{}
		m, id, err = repo.model.(Mapable).Mapper().Pack(columns, cols, col)
		if err != nil {
			return
		}
		models[id] = m
	})
	if err != nil {
		return
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
		row = repo.model.(Mapable).Mapper().Extract(models)
		repo.onupdate(models)
		_, err = repo.exec(repo.ForUpdate(row))
		return err
	case reflect.Slice:
		slice := models.([]interface{})
		for _, m := range slice {
			row = repo.model.(Mapable).Mapper().Extract(m)
			repo.onupdate(m)
			_, err = repo.exec(repo.ForUpdate(row))
			return err
		}
	case reflect.Map:
		maps := models.([]interface{})
		for _, m := range maps {
			row = repo.model.(Mapable).Mapper().Extract(m)
			repo.onupdate(m)
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
		repo.oncreate(val.Interface())
		data = append(data, repo.model.(Mapable).Mapper().Extract(val.Interface()))
	case reflect.Slice:
		slice := val.Interface().([]interface{})
		for _, m := range slice {
			repo.oncreate(m)
			data = append(data, repo.model.(Mapable).Mapper().Extract(m))
		}
	case reflect.Map:
		maps := val.Interface().(map[interface{}]interface{})
		for _, m := range maps {
			repo.oncreate(m)
			data = append(data, repo.model.(Mapable).Mapper().Extract(m))
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
