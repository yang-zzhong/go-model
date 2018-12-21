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
	modifier Modifier    // sql modifier
	oncreate modify      // on create callback
	onupdate modify      // on update callback
	ondelete modify      // on delete callback
	withs    []with      // maintain fetch model relationship
	*Builder
}

// new custom repo
func NewCustomRepo(m interface{}, p Modifier) *Repo {
	repo := new(Repo)
	repo.model = m
	repo.modifier = p
	repo.oncreate = func(_ interface{}) error { return nil }
	repo.onupdate = func(_ interface{}) error { return nil }
	repo.ondelete = func(_ interface{}) error { return nil }
	repo.Builder = NewBuilder(p)
	repo.withs = []with{}
	repo.From(repo.model.(Model).TableName())

	return repo
}

// new default repo
func NewRepo(m interface{}) (repo *Repo, err error) {
	repo = NewCustomRepo(m, modifier)
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
	return repo.CountDB(GetDB())
}

func (repo *Repo) CountDB(db DB) (int, error) {
	rows, err := db.Query(repo.ForCount(), repo.Params()...)
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
	return repo.QueryDB(handle, GetDB())
}

func (repo *Repo) QueryDB(handle rowshandler, db DB) error {
	rows, err := db.Query(repo.ForCount(), repo.Params()...)
	if err != nil {
		return err
	}
	defer rows.Close()
	var cols []string
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

func (repo *Repo) UpdateSlice(models []interface{}, db DB) error {
	mm := repo.model.(Mapable).Mapper()
	field := repo.model.(Model).PK()
	ids := []interface{}{}
	for _, m := range models {
		v, err := mm.colValue(m, field)
		if err != nil {
			return err
		}
		ids = append(ids, v)
		if err := repo.onupdate(m); err != nil {
			return err
		}
	}
	if len(ids) > 0 {
		repo.WhereIn(field, ids)
		if err := r.UpdateRaw(mm.extract(m)); err != nil {
			return err
		}
	}

	return nil
}

func (repo *Repo) Update(model interface{}) error {
	v, err := mm.colValue(model, field)
	if err != nil {
		return err
	}
	if err := repo.onupdate(model); err != nil {
		return err
	}
	r.Where(field, v)

	return r.UpdateRaw(mm.extract(model))
}

func (repo *Repo) CreateSlice(models []interface{}) error {
	return repo.CreateSliceDB(models, GetDB())
}

func (repo *Repo) CreateSliceDB(models []interface{}, db DB) error {
	r := NewCustomRepo(repo.model, repo.modifier)
	var data []map[string]interface{}
	for _, m := range models {
		if err := repo.oncreate(m); err != nil {
			return err
		}
		data = append(data, repo.model.(Mapable).Mapper().extract(m))
	}
	if _, err := db.Exec(r.ForInsert(data), r.Params()...); err != nil {
		return err
	}
	for _, m := range models {
		m.(Model).SetFresh(false)
	}

	return nil
}

func (repo *Repo) Create(model interface{}) error {
	return repo.CreateDB(model, GetDB())
}

func (repo *Repo) CreateDB(model interface{}, db DB) error {
	var data []map[string]interface{}
	if err := repo.oncreate(models); err != nil {
		return err
	}
	data = append(data, repo.model.(Mapable).Mapper().extract(model))
	r := NewCustomRepo(repo.model, repo.modifier)
	if _, err := db.Exec(r.ForInsert(data), r.Params()...); err != nil {
		return err
	}
	m.(Model).SetFresh(false)

	return nil
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
	// defer func() {
	// 	if e := recover(); e != nil {
	// 		switch e.(type) {
	// 		case string:
	// 			err = &Error{ERR_SQL, errors.New(e.(string))}
	// 		case error:
	// 			err = &Error{ERR_SQL, e.(error)}
	// 		}
	// 	}
	// }()
	// if repo.tx != nil {
	// 	if r, e := repo.tx.Exec(sql, repo.Params()...); e != nil {
	// 		err = &Error{ERR_SQL, e}
	// 	} else {
	// 		res = r
	// 	}
	// } else {
	// 	if r, e := repo.db.Exec(sql, repo.Params()...); e != nil {
	// 		err = &Error{ERR_SQL, e}
	// 	} else {
	// 		res = r
	// 	}
	// }
	// return
}

func (repo *Repo) query() (rows *sql.Rows, err error) {
	// defer func() {
	// 	if e := recover(); e != nil {
	// 		switch e.(type) {
	// 		case string:
	// 			err = &Error{ERR_SQL, errors.New(e.(string))}
	// 		case error:
	// 			err = &Error{ERR_SQL, e.(error)}
	// 		}
	// 	}
	// }()
	// if r, e := repo.conn.Query(repo.ForQuery(), repo.Params()...); e != nil {
	// 	err = &Error{ERR_SQL, e}
	// } else {
	// 	rows = r
	// }

	// return
}
