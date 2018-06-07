package model

import (
	"context"
	"database/sql"
	"errors"
	. "github.com/yang-zzhong/go-querybuilder"
	"reflect"
)

type rowshandler func(*sql.Rows, []string)

// oncreate and onupdate callback type
type modify func(model interface{})
type setpage func(*Repo) error

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
	t_bad  = 3 // bad or not found relationship
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
	ondelete modify      // on delete callback
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

// set on update callback
func (repo *Repo) OnDelete(c modify) *Repo {
	repo.ondelete = c
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

func (repo *Repo) FetchPage(handle setpage) (models map[interface{}]interface{}, total int, err error) {
	if total, err = repo.Count(); err != nil {
		return
	}
	if err = handle(repo); err != nil {
		return
	}
	models, err = repo.Fetch()
	return
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
	if nexusValues, rerr := repo.nexusValues(models); rerr == nil {
		for id, _ := range models {
			repo.bindNexus(models[id], nexusValues)
		}
	} else {
		err = rerr
	}

	return
}

func (repo *Repo) One() (interface{}, bool, error) {
	if rows, err := repo.Fetch(); err != nil {
		return nil, false, err
	} else {
		for _, row := range rows {
			return row, true, nil
		}
	}

	return nil, false, nil
}

func (repo *Repo) Find(id interface{}) (interface{}, bool, error) {
	r := NewCustomRepo(repo.model, repo.conn, repo.modifier)
	r.Where(repo.model.(Model).PK(), id).Limit(1)
	if rows, err := r.Fetch(); err != nil {
		return nil, false, err
	} else {
		for _, row := range rows {
			return row, true, nil
		}
	}

	return nil, false, nil
}

func (repo *Repo) UpdateRaw(raw map[string]interface{}) error {
	_, err := repo.exec(repo.ForUpdate(raw))
	return err
}

func (repo *Repo) DeleteRaw() error {
	_, err := repo.exec(repo.ForRemove())
	return err
}

func (repo *Repo) Update(models interface{}) error {
	val := reflect.ValueOf(models)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	r := NewCustomRepo(repo.model, repo.conn, repo.modifier)
	field := repo.model.(Model).PK()
	mm := repo.model.(Mapable).Mapper()
	switch val.Kind() {
	case reflect.Struct:
		if v, err := mm.ColValue(models, field); err == nil {
			repo.onupdate(models)
			r.Where(field, v)
			return r.UpdateRaw(mm.Extract(models))
		} else {
			return err
		}
	case reflect.Slice:
		slice := models.([]interface{})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		return r.Tx(func(tx *sql.Tx) error {
			r.WithTx(tx)
			for _, m := range slice {
				if v, err := mm.ColValue(m, field); err == nil {
					repo.onupdate(m)
					r.Where(field, v)
					if err := r.UpdateRaw(mm.Extract(m)); err != nil {
						return err
					}
				} else {
					return err
				}
			}
			return nil
		}, ctx, nil)
	case reflect.Map:
		maps := models.([]interface{})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		return r.Tx(func(tx *sql.Tx) error {
			r.WithTx(tx)
			for _, m := range maps {
				if v, err := mm.ColValue(m, field); err == nil {
					repo.onupdate(m)
					r.Where(field, v)
					if err := r.UpdateRaw(mm.Extract(m)); err != nil {
						return err
					}
				} else {
					return err
				}
			}
			return nil
		}, ctx, nil)
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
	ms := []interface{}{}
	switch val.Kind() {
	case reflect.Struct:
		repo.oncreate(models)
		data = append(data, repo.model.(Mapable).Mapper().Extract(models))
		ms = append(ms, models)
	case reflect.Slice:
		slice := val.Interface().([]interface{})
		for _, m := range slice {
			repo.oncreate(m)
			data = append(data, repo.model.(Mapable).Mapper().Extract(m))
			ms = append(ms, m)
		}
	case reflect.Map:
		maps := val.Interface().(map[interface{}]interface{})
		for _, m := range maps {
			repo.oncreate(m)
			data = append(data, repo.model.(Mapable).Mapper().Extract(m))
			ms = append(ms, m)
		}
	}
	r := NewCustomRepo(repo.model, repo.conn, repo.modifier)
	_, err = r.exec(r.ForInsert(data))
	if err == nil {
		for _, m := range ms {
			m.(BaseI).SetFresh(false)
		}
	}

	return err
}

func (repo *Repo) Delete(models interface{}) error {
	val := reflect.ValueOf(models)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	r := NewCustomRepo(repo.model, repo.conn, repo.modifier)
	field := r.model.(Model).PK()
	ins := []interface{}{}
	mm := r.model.(Mapable).Mapper()
	ms := []interface{}{}
	switch val.Kind() {
	case reflect.Struct:
		repo.ondelete(val.Interface())
		if v, err := mm.ColValue(models, field); err == nil {
			ins = append(ins, v)
			ms = append(ms, models)
		} else {
			return err
		}
	case reflect.Slice:
		slice := val.Interface().([]interface{})
		for _, m := range slice {
			repo.ondelete(m)
			if v, err := mm.ColValue(m, field); err == nil {
				ins = append(ins, v)
				ms = append(ms, m)
			} else {
				return err
			}
		}
	case reflect.Map:
		maps := val.Interface().([]interface{})
		for _, m := range maps {
			repo.ondelete(m)
			if v, err := mm.ColValue(m, field); err == nil {
				ins = append(ins, v)
				ms = append(ms, m)
			} else {
				return err
			}
		}
	}
	if len(ins) == 0 {
		return nil
	}
	r.WhereIn(field, ins)
	err := r.DeleteRaw()
	if err == nil {
		for _, m := range ms {
			m.(BaseI).SetFresh(true)
		}
	}

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
