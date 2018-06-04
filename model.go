package model

import (
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
	SetMany(name string, many map[interface{}]interface{})
}

type ValueConverter interface {
	DBValue(fieldName string, value interface{}) interface{}
	Value(fieldName string, value interface{}) (reflect.Value, bool)
}

type RelationShip struct {
	target interface{}
	n      Nexus
}

type Base struct {
	ones       map[string]RelationShip
	manys      map[string]RelationShip
	onesValue  map[string]interface{}
	manysValue map[string]map[interface{}]interface{}
}

func NewBase() *Base {
	m := new(Base)
	m.ones = make(map[string]RelationShip)
	m.manys = make(map[string]RelationShip)
	m.onesValue = make(map[string]interface{})
	m.manysValue = make(map[string]map[interface{}]interface{})
	return m
}

func (m *Base) DeclareOne(name string, one interface{}, n Nexus) {
	m.ones[name] = RelationShip{one, n}
}

func (m *Base) DeclareMany(name string, many interface{}, n Nexus) {
	m.manys[name] = RelationShip{many, n}
}

func (m *Base) HasOne(name string) (one interface{}, n Nexus, has bool) {
	if conf, ok := m.ones[name]; ok {
		one = conf.target
		n = conf.n
		has = true
		return
	}
	has = false
	return
}

func (m *Base) HasMany(name string) (many interface{}, n Nexus, has bool) {
	if conf, ok := m.manys[name]; ok {
		many = conf.target
		n = conf.n
		has = true
		return
	}
	has = false
	return
}

func (m *Base) SetOne(name string, model interface{}) {
	m.onesValue[name] = model
}

func (m *Base) SetMany(name string, models map[interface{}]interface{}) {
	m.manysValue[name] = models
}

func (m *Base) findOne(model interface{}, name string) (result interface{}, err error) {
	var one interface{}
	var n Nexus
	var has bool
	if one, n, has = m.HasOne(name); !has {
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

func (m *Base) findMany(model interface{}, name string) (result map[interface{}]interface{}, err error) {
	var many interface{}
	var rel map[string]string
	var has bool
	if many, rel, has = m.HasMany(name); !has {
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

func One(m *Base, model interface{}, name string) (one interface{}, err error) {
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

func Many(m *Base, model interface{}, name string) (many map[interface{}]interface{}, err error) {
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
