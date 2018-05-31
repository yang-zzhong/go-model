package model

import (
	"errors"
	"reflect"
)

type Model interface {
	TableName() string // table name in db
	PK() string        // primary key
	HasOne(name string) (Model, map[string]string, error)
	HasMany(name string) (Model, map[string]string, error)
}

type ValueConverter interface {
	DBValue(fieldName string, value interface{}) interface{}
	Value(fieldName string, value interface{}) (reflect.Value, bool)
}

type Relation struct {
	rel    Model
	linker map[string]string
}

type BaseModel struct {
	hasOne  map[string]Relation
	hasMany map[string]Relation
}

func (m *BaseModel) DeclareOne(name string, one Model, relation map[string]string) {
	m.hasOne[name] = Relation{one, relation}
}

func (m *BaseModel) DeclareMany(name string, many Model, relation map[string]string) {
	m.hasMany[name] = Relation{many, relation}
}

func (m *BaseModel) HasOne(name string) (one Model, rel map[string]string, err error) {
	if conf, ok := m.hasOne[name]; ok {
		one = conf.rel
		rel = one.linker
		return
	}
	err = errors.New("relation " + name + " not found!")
	return
}

func (m *BaseModel) HasMany(name string) (one Model, rel map[string]string, err error) {
	if conf, ok := m.hasMany[name]; ok {
		one = conf.rel
		rel = one.linker
		return
	}
	err = errors.New("relation " + name + " not found!")
	return
}

func (m *BaseModel) One(name string) (result interface{}, err error) {
	var one interface{}
	var rel map[string]string
	if one, rel, err = m.HasOne(name); err != nil {
		return
	}
	var repo *Repo
	if repo, err = NewRepo(one); err != nil {
		return
	}
	for af, bf := range rel {
		repo.Where(bf, querybuilder.Field(af))
	}
	result, err = repo.One()

	return
}

func (m *BaseModel) Many(name string) (result map[string]interface{}, err error) {
	var many interface{}
	var rel map[string]string
	if many, rel, err = m.HasMany(name); err != nil {
		return
	}
	var repo *Repo
	if repo, err = NewRepo(many); err != nil {
		return
	}
	for af, bf := range rel {
		repo.Where(bf, querybuilder.Field(af))
	}
	result, err = repo.Fetch()

	return
}
