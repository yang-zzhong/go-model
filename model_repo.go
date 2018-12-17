package model

import (
	"database/sql"
	"errors"
	. "github.com/yang-zzhong/go-querybuilder"
	"reflect"
)

type rowshandler func(*sql.Rows, []string) error
type handlerForQueryModel func(m interface{}, pk interface{}) error

// oncreate and onupdate callback type
type modify func(model interface{}) error
type setpage func(*Repo) error

const (
	t_one  = 1 // relationship is a has one relationship
	t_many = 2 // relationship is a has many relationship
	t_bad  = 3 // bad or not found relationship
)

// a relationship with repo
type with struct {
	name    string      // relationship name
	m       interface{} // relationship target
	n       Nexus       // relationship nexus
	t       int         // relationship type t_one|t_many
	handler repoHandler
}

// repo
type Repo struct {
	model    interface{} // repo row model
	conn     *Connection // db
	modifier Modifier    // sql modifier
	oncreate modify      // on create callback
	onupdate modify      // on update callback
	ondelete modify      // on delete callback
	withs    []with      // maintain fetch model relationship
	tx       *sql.Tx     // tx
	*Builder
}

// new custom repo
func NewCustomRepo(m interface{}, conn *Connection, p Modifier) *Repo {
	repo := new(Repo)
	repo.model = m
	repo.conn = conn
	repo.modifier = p
	repo.oncreate = func(_ interface{}) error { return nil }
	repo.onupdate = func(_ interface{}) error { return nil }
	repo.ondelete = func(_ interface{}) error { return nil }
	repo.Builder = NewBuilder(p)
	repo.withs = []with{}
	repo.From(repo.model.(Model).TableName())

	return repo
}

func (repo *Repo) WithTx(tx *sql.Tx) *Repo {
	repo.tx = tx

	return repo
}

func (repo *Repo) WithoutTx() *Repo {
	repo.tx = nil

	return repo
}

// new default repo
func NewRepo(m interface{}) (repo *Repo, err error) {
	if !inited {
		err = errors.New("not config the db and modifier yet")
		return
	}
	repo = NewCustomRepo(m, Conn, modifier)
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
	defer rows.Close()
	var count int
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return 0, err
		}
	}
	return count, nil
}

func (repo *Repo) MustCount() int {
	if count, err := repo.Count(); err != nil {
		panic(err)
	} else {
		return count
	}
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
		if err := handle(rows, cols); err != nil {
			return err
		}
	}

	return nil
}

func (repo *Repo) MustFetch() []interface{} {
	if ms, err := repo.Fetch(); err != nil {
		panic(err)
	} else {
		return ms
	}
}

func (repo *Repo) Fetch() (models []interface{}, err error) {
	err = repo.fetch(func(m interface{}, _ interface{}) error {
		models = append(models, m)
		return nil
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

func (repo *Repo) MustFetchKey(col string) map[interface{}]interface{} {
	if ms, err := repo.FetchKey(col); err != nil {
		panic(err)
	} else {
		return ms
	}
}

func (repo *Repo) FetchKey(col string) (models map[interface{}]interface{}, err error) {
	models = make(map[interface{}]interface{})
	forNexusValues := []interface{}{}
	err = repo.fetch(func(m interface{}, id interface{}) error {
		models[id] = m
		forNexusValues = append(forNexusValues, m)
		return nil
	})
	if err != nil {
		return
	}
	if nexusValues, rerr := repo.nexusValues(forNexusValues); rerr == nil {
		for id, _ := range models {
			repo.bindNexus(models[id], nexusValues)
		}
	} else {
		err = rerr
	}

	return
}

func (repo *Repo) fetch(handle handlerForQueryModel) error {
	var cols []interface{}
	colget := false
	return repo.Query(func(rows *sql.Rows, columns []string) error {
		var err error
		if !colget {
			if cols, err = repo.model.(Mapable).Mapper().cols(columns); err != nil {
				return err
			}
			colget = true
		}
		if err = rows.Scan(cols...); err != nil {
			return &Error{ERR_SCAN, err}
		}
		var m, id interface{}
		m, id, err = repo.model.(Mapable).Mapper().pack(columns, cols, repo.model.(Model).PK())
		if err == nil {
			return handle(m, id)
		}
		return err
	})
}

func (repo *Repo) MustOne() interface{} {
	if m, exist, err := repo.One(); err != nil {
		panic(err)
	} else if !exist {
		panic(&Error{ERR_DATA_NOT_FOUND, errors.New("data not found")})
	} else {
		return m
	}
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

func (repo *Repo) MustFind(id interface{}) interface{} {
	if m, exist, err := repo.Find(id); err != nil {
		panic(err)
	} else if !exist {
		panic(&Error{ERR_DATA_NOT_FOUND, errors.New("data not found")})
	} else {
		return m
	}
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
	r.tx = repo.tx
	field := repo.model.(Model).PK()
	mm := repo.model.(Mapable).Mapper()
	switch val.Kind() {
	case reflect.Struct:
		if v, err := mm.colValue(models, field); err == nil {
			if err := repo.onupdate(models); err != nil {
				return err
			}
			r.Where(field, v)
			return r.UpdateRaw(mm.extract(models))
		} else {
			return err
		}
	case reflect.Slice:
		slice := models.([]interface{})
		for _, m := range slice {
			if v, err := mm.colValue(m, field); err == nil {
				if err := repo.onupdate(m); err != nil {
					return err
				}
				r.Where(field, v)
				if err := r.UpdateRaw(mm.extract(m)); err != nil {
					return err
				}
			} else {
				return err
			}
		}
		return nil
	case reflect.Map:
		maps := models.([]interface{})
		for _, m := range maps {
			if v, err := mm.colValue(m, field); err == nil {
				if err := repo.onupdate(m); err != nil {
					return err
				}
				r.Where(field, v)
				if err := r.UpdateRaw(mm.extract(m)); err != nil {
					return err
				}
			} else {
				return err
			}
		}
		return nil
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
		if err := repo.oncreate(models); err != nil {
			return err
		}
		data = append(data, repo.model.(Mapable).Mapper().extract(models))
		ms = append(ms, models)
	case reflect.Slice:
		slice := val.Interface().([]interface{})
		for _, m := range slice {
			if err := repo.oncreate(m); err != nil {
				return err
			}
			data = append(data, repo.model.(Mapable).Mapper().extract(m))
			ms = append(ms, m)
		}
	case reflect.Map:
		maps := val.Interface().(map[interface{}]interface{})
		for _, m := range maps {
			if err := repo.oncreate(m); err != nil {
				return err
			}
			data = append(data, repo.model.(Mapable).Mapper().extract(m))
			ms = append(ms, m)
		}
	}
	r := NewCustomRepo(repo.model, repo.conn, repo.modifier)
	r.tx = repo.tx
	_, err = r.exec(r.ForInsert(data))
	if err == nil {
		for _, m := range ms {
			m.(Model).SetFresh(false)
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
	r.tx = repo.tx
	field := r.model.(Model).PK()
	ins := []interface{}{}
	mm := r.model.(Mapable).Mapper()
	ms := []interface{}{}
	switch val.Kind() {
	case reflect.Struct:
		if err := repo.ondelete(models); err != nil {
			return err
		}
		if v, err := mm.colValue(models, field); err == nil {
			ins = append(ins, v)
			ms = append(ms, models)
		} else {
			return err
		}
	case reflect.Slice:
		slice := val.Interface().([]interface{})
		for _, m := range slice {
			if err := repo.ondelete(m); err != nil {
				return err
			}
			if v, err := mm.colValue(m, field); err == nil {
				ins = append(ins, v)
				ms = append(ms, m)
			} else {
				return err
			}
		}
	case reflect.Map:
		maps := val.Interface().([]interface{})
		for _, m := range maps {
			if err := repo.ondelete(m); err != nil {
				return err
			}
			if v, err := mm.colValue(m, field); err == nil {
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
			m.(Model).SetFresh(true)
		}
	}

	return err
}

func (repo *Repo) exec(sql string) (res sql.Result, err error) {
	defer func() {
		if e := recover(); e != nil {
			switch e.(type) {
			case string:
				err = &Error{ERR_SQL, errors.New(e.(string))}
			case error:
				err = &Error{ERR_SQL, e.(error)}
			}
		}
	}()
	if repo.tx != nil {
		if r, e := repo.tx.Exec(sql, repo.Params()...); e != nil {
			err = &Error{ERR_SQL, e}
		} else {
			res = r
		}
	} else {
		if r, e := repo.conn.Exec(sql, repo.Params()...); e != nil {
			err = &Error{ERR_SQL, e}
		} else {
			res = r
		}
	}
	return
}

func (repo *Repo) query() (rows *sql.Rows, err error) {
	defer func() {
		if e := recover(); e != nil {
			switch e.(type) {
			case string:
				err = &Error{ERR_SQL, errors.New(e.(string))}
			case error:
				err = &Error{ERR_SQL, e.(error)}
			}
		}
	}()
	if r, e := repo.conn.Query(repo.ForQuery(), repo.Params()...); e != nil {
		err = &Error{ERR_SQL, e}
	} else {
		rows = r
	}

	return
}
