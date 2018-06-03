package model

import (
	"errors"
	"reflect"
)

type Nexus map[string]string

type Model interface {
	//
	// table name in database
	//
	TableName() string
	//
	// primary key for the table
	//
	PK() string
}

type NexusOne interface {
	HasOne(name string) (interface{}, Nexus, bool)
	DeclareOne(name string, one interface{}, n Nexus)
	SetOne(name string, one interface{})
}

type NexusMany interface {
	HasMany(name string) (interface{}, Nexus, bool)
	DeclareMany(name string, many interface{}, n Nexus)
	SetMany(name string, many interface{})
}

type ValueConverter interface {
	DBValue(fieldName string, value interface{}) interface{}
	Value(fieldName string, value interface{}) (reflect.Value, bool)
}

type RelationShip struct {
	target interface{}
	n      Nexus
}

type BaseModel struct {
	ones       map[string]RelationShip
	manys      map[string]RelationShip
	onesValue  map[string]interface{}
	manysValue map[string]map[interface{}]interface{}
}

func NewBaseModel() *BaseModel {
	m := new(BaseModel)
	m.ones = make(map[string]RelationShip)
	m.manys = make(map[string]RelationShip)
	m.onesValue = make(map[string]interface{})
	m.manysValue = make(map[string]map[interface{}]interface{})
	return m
}

func (m *BaseModel) DeclareOne(name string, one interface{}, n Nexus) {
	m.ones[name] = RelationShip{one, n}
}

func (m *BaseModel) DeclareMany(name string, many interface{}, n Nexus) {
	m.manys[name] = RelationShip{many, n}
}

func (m *BaseModel) HasOne(name string) (one interface{}, n Nexus, err error) {
	if conf, ok := m.ones[name]; ok {
		one = conf.target
		n = conf.n
		return
	}
	err = errors.New("relationship " + name + " not found!")
	return
}

func (m *BaseModel) HasMany(name string) (many interface{}, n Nexus, err error) {
	if conf, ok := m.manys[name]; ok {
		many = conf.target
		n = conf.n
		return
	}
	err = errors.New("relationship " + name + " not found!")
	return
}

func (m *BaseModel) SetOne(name string, model interface{}) {
	m.onesValue[name] = model
}

func (m *BaseModel) SetMany(name string, models map[interface{}]interface{}) {
	m.manysValue[name] = models
}

func (m *BaseModel) findOne(model interface{}, name string) (result interface{}, err error) {
	var one interface{}
	var n Nexus
	if one, n, err = m.HasOne(name); err != nil {
		return
	}
	var repo *Repo
	if repo, err = NewRepo(one); err != nil {
		return
	}
	for af, bf := range n {
		value, err := fieldValue(model, af)
		if err != nil {
			return result, err
		}
		repo.Where(bf, value)
	}
	result, err = repo.One()

	return
}

func (m *BaseModel) findMany(model interface{}, name string) (result map[interface{}]interface{}, err error) {
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
		value, err := fieldValue(model, af)
		if err != nil {
			return result, err
		}
		repo.Where(bf, value)
	}
	result, err = repo.Fetch()

	return
}

func One(m *BaseModel, model interface{}, name string) (one interface{}, err error) {
	if v, ok := m.onesValue[name]; ok {
		one = v
		return
	}
	if m.onesValue[name], err = m.findOne(model, name); err != nil {
		return
	}
	one = m.onesValue[name]
	return
}

func Many(m *BaseModel, model interface{}, name string) (many map[interface{}]interface{}, err error) {
	if v, ok := m.manysValue[name]; ok {
		many = v
		return
	}
	if m.manysValue[name], err = m.findMany(model, name); err != nil {
		return
	}
	many = m.manysValue[name]
	return
}

func fieldValue(m interface{}, field string) (value interface{}, err error) {
	mm := NewModelMapper(m)
	value, err = mm.ColValue(m, field)

	return
}
