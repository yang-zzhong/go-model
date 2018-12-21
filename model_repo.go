package model

import (
	"database/sql"
	"errors"
	. "github.com/yang-zzhong/go-querybuilder"
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
	db       DB
	modifier Modifier // sql modifier
	oncreate modify   // on create callback
	onupdate modify   // on update callback
	ondelete modify   // on delete callback
	withs    []with   // maintain fetch model relationship
	*Builder
}

// new custom repo
func NewRepo(m interface{}, db DB, p Modifier) *Repo {
	repo := new(Repo)
	repo.model = m
	repo.db = db
	repo.modifier = p
	repo.oncreate = func(_ interface{}) error { return nil }
	repo.onupdate = func(_ interface{}) error { return nil }
	repo.ondelete = func(_ interface{}) error { return nil }
	repo.Builder = NewBuilder(p)
	repo.withs = []with{}
	repo.From(repo.model.(Model).TableName())

	return repo
}

func (repo *Repo) Another() *Repo {
	r := new(Repo)
	r.model = repo.model
	r.db = repo.db
	r.modifier = repo.modifier
	r.oncreate = repo.oncreate
	r.onupdate = repo.onupdate
	r.ondelete = repo.ondelete
	r.Builder = NewBuilder(r.modifier)
	r.withs = []with{}
	r.From(r.model.(Model).TableName())

	return r
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

func (repo *Repo) Count() (int, error) {
	rows, err := repo.db.Query(repo.ForCount(), repo.Params()...)
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
	return repo.QueryDB(handle, GetDefaultDB())
}

func (repo *Repo) QueryDB(handle rowshandler, db DB) error {
	rows, err := db.Query(repo.ForQuery(), repo.Params()...)
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
	r := repo.Another()
	pk := repo.model.(Model).PK()
	r.Where(pk, id).Limit(1)
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
	_, err := repo.db.Exec(repo.ForUpdate(raw), repo.Params()...)
	return err
}

func (repo *Repo) DeleteRaw(raw map[string]interface{}) error {
	_, err := repo.db.Exec(repo.ForRemove(), repo.Params()...)
	return err
}

func (repo *Repo) Update(model interface{}) error {
	field := repo.model.(Model).PK()
	v, err := repo.model.(Mapable).Mapper().colValue(model, field)
	if err != nil {
		return err
	}
	if err := repo.onupdate(model); err != nil {
		return err
	}
	r := repo.Another()
	sql := r.Where(field, v).ForUpdate(repo.model.(Mapable).Mapper().extract(model))
	_, err = repo.db.Exec(sql, r.Params()...)

	return err
}

func (repo *Repo) CreateSlice(models []interface{}) error {
	r := repo.Another()
	var data []map[string]interface{}
	for _, m := range models {
		if err := r.oncreate(m); err != nil {
			return err
		}
		data = append(data, r.model.(Mapable).Mapper().extract(m))
	}
	if _, err := repo.db.Exec(r.ForInsert(data), r.Params()...); err != nil {
		return err
	}
	for _, m := range models {
		m.(Model).SetFresh(false)
	}

	return nil
}

func (repo *Repo) Create(model interface{}) error {
	var data []map[string]interface{}
	if err := repo.oncreate(model); err != nil {
		return err
	}
	data = append(data, repo.model.(Mapable).Mapper().extract(model))
	r := repo.Another()
	if _, err := repo.db.Exec(r.ForInsert(data), r.Params()...); err != nil {
		return err
	}
	model.(Model).SetFresh(false)

	return nil
}

func (repo *Repo) Delete(model interface{}) error {
	if err := repo.ondelete(model); err != nil {
		return err
	}
	field := repo.model.(Model).PK()
	v, err := repo.model.(Mapable).Mapper().colValue(model, field)
	if err != nil {
		return err
	}
	r := repo.Another()
	_, err = r.db.Exec(r.Where(field, v).ForRemove(), r.Params()...)
	return err
}

func (repo *Repo) DeleteSlice(models []interface{}) error {
	field := repo.model.(Model).PK()
	ids := []interface{}{}
	for _, model := range models {
		v, err := repo.model.(Mapable).Mapper().colValue(model, field)
		if err != nil {
			return err
		}
		ids = append(ids, v)
	}
	r := repo.Another()
	_, err := r.db.Exec(r.WhereIn(field, ids).ForRemove(), r.Params()...)
	return err
}
