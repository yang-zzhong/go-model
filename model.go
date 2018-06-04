package model

import (
	"reflect"
)

//
// column to column relationship
//
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

//
// a has one nexus
//
type NexusOne interface {
	// get a has one nexus from it's name
	HasOne(name string) (interface{}, Nexus, bool)
	// declare a has one relationship
	DeclareOne(name string, one interface{}, n Nexus)
	// set fetch result of has one relationship
	SetOne(name string, one interface{})
}

//
// a has many nexus
//
type NexusMany interface {
	// get a has many nexus from it's name
	HasMany(name string) (interface{}, Nexus, bool)
	// declare a has many relationship
	DeclareMany(name string, many interface{}, n Nexus)
	// set fetch result of has many relationship
	SetMany(name string, many map[interface{}]interface{})
}

//
// value converter from database value to struct field value and from struct filed value to database value
//
type ValueConverter interface {
	// convert value to database value
	DBValue(fieldName string, value interface{}) interface{}
	// conver database value to struct field value
	Value(fieldName string, value interface{}) (reflect.Value, bool)
}

// relationship
type relationship struct {
	target interface{} // related with who
	n      Nexus       // related
}

// base model struct
type Base struct {
	ones       map[string]relationship                // has one relationship
	manys      map[string]relationship                // has many relationship
	onesValue  map[string]interface{}                 // fetched result of has one relationship
	manysValue map[string]map[interface{}]interface{} // fetched result of has many relationship
}

// new a base model
func NewBase() *Base {
	m := new(Base)
	m.ones = make(map[string]relationship)
	m.manys = make(map[string]relationship)
	m.onesValue = make(map[string]interface{})
	m.manysValue = make(map[string]map[interface{}]interface{})
	return m
}

func (m *Base) DeclareOne(name string, one interface{}, n Nexus) {
	m.ones[name] = relationship{one, n}
}

func (m *Base) DeclareMany(name string, many interface{}, n Nexus) {
	m.manys[name] = relationship{many, n}
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
