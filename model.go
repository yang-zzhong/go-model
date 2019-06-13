package model

import (
	"encoding/json"
	"errors"
	"reflect"
)

// column to column relationship
type Nexus map[string]interface{}

type Preparable interface {
	Prepare()
}

type Model interface {
	TableName() string // table name in database
	PK() string        // primary key for the table
	Repo() *Repo
	Save() error                            // save to db
	Fill(data map[string]interface{})       // fill values
	Set(name string, val interface{}) error // set col value
	Has(name string) bool
	Get(name string) interface{} // set col value
	SetFresh(fresh bool)
	DB() *Db
	IsFresh() bool
	OnCreate(handle modify)
	OnUpdate(handle modify)
	OnDelete(handle modify)
}

// a has one nexus
type NexusOne interface {
	HasOne(name string) (interface{}, Nexus, bool)    // get a has one nexus from it's name
	DeclareOne(name string, one interface{}, n Nexus) // declare a has one relationship
	SetOne(name string, one interface{})              // set fetch result of has one relationship
}

// a has many nexus
type NexusMany interface {
	HasMany(name string) (interface{}, Nexus, bool)     // get a has many nexus from it's name
	DeclareMany(name string, many interface{}, n Nexus) // declare a has many relationship
	SetMany(name string, many interface{})              // set fetch result of has many relationship
}

type Mapable interface {
	Mapper() *ModelMapper
}

// value converter from database value to struct field value and from struct filed value to database value
type ValueConverter interface {
	DBValue(fieldName string, value interface{}) interface{}         // convert value to database value
	Value(fieldName string, value interface{}) (reflect.Value, bool) // conver database value to struct field value
}

// relationship
type relationship struct {
	target interface{} // related with who
	n      Nexus       // related
}

// base model struct
type Base struct {
	mapper     *ModelMapper
	fresh      bool
	repo       *Repo
	dbSelector string
	ones       map[string]relationship // has one relationship
	manys      map[string]relationship // has many relationship
	onesValue  map[string]interface{}  // fetched result of has one relationship
	manysValue map[string]interface{}  // fetched result of has many relationship
	oncreate   modify                  // declare in model_repo.go
	onupdate   modify
	ondelete   modify
}

// new a base model
func newBase(m interface{}) *Base {
	base := new(Base)
	base.fresh = true
	base.dbSelector = DEFAULT
	base.mapper = NewModelMapper(m)
	base.ones = make(map[string]relationship)
	base.manys = make(map[string]relationship)
	base.onesValue = make(map[string]interface{})
	base.manysValue = make(map[string]interface{})
	return base
}

func (base *Base) OnCreate(m modify) {
	base.Repo().oncreate = m
}

func (base *Base) DB() *Db {
	return GetDB(base.DBSelector())
}

func (base *Base) DBSelector() string {
	return base.dbSelector
}

func (base *Base) SetDBSelector(selector string) {
	base.dbSelector = selector
}

func (base *Base) OnUpdate(m modify) {
	base.Repo().onupdate = m
}

func (base *Base) OnDelete(m modify) {
	base.Repo().ondelete = m
}

func (base *Base) IsFresh() bool {
	return base.fresh
}

func (base *Base) SetFresh(fresh bool) {
	base.fresh = fresh
}

func (m *Base) DeclareOne(name string, one interface{}, n Nexus) {
	m.ones[name] = relationship{one, n}
}

func (base *Base) DeclareMany(name string, many interface{}, n Nexus) {
	base.manys[name] = relationship{many, n}
}

func (base *Base) HasOne(name string) (one interface{}, n Nexus, has bool) {
	if conf, ok := base.ones[name]; ok {
		one = conf.target
		n = conf.n
		has = true
		return
	}
	has = false
	return
}

func (base *Base) HasMany(name string) (many interface{}, n Nexus, has bool) {
	if conf, ok := base.manys[name]; ok {
		many = conf.target
		n = conf.n
		has = true
		return
	}
	has = false
	return
}

func (base *Base) Mapper() *ModelMapper {
	return base.mapper
}

func (base *Base) SetOne(name string, model interface{}) {
	base.onesValue[name] = model
}

func (base *Base) SetMany(name string, models interface{}) {
	base.manysValue[name] = models
}

func (base *Base) findOne(name string) (result interface{}, err error) {
	var one interface{}
	var n Nexus
	var has bool
	if one, n, has = base.HasOne(name); !has {
		err = errors.New("nexus of " + name + "does not exist")
		return
	}
	m := New(one).(Model)
	repo := m.Repo()
	for af, bf := range n {
		switch bf.(type) {
		case NWhere:
			repo.Where(af, bf.(NWhere).Op, bf.(NWhere).Value)
		case string:
			value, err := base.fieldValue(bf.(string))
			if err != nil {
				return result, err
			}
			repo.Where(af, value)
		}
	}
	result, _, err = repo.One()

	return
}

func (base *Base) findMany(name string) (result interface{}, err error) {
	var many interface{}
	var rel Nexus
	var has bool
	if many, rel, has = base.HasMany(name); !has {
		return
	}
	m := New(many).(Model)
	repo := m.Repo()
	for af, bf := range rel {
		switch bf.(type) {
		case NWhere:
			repo.Where(af, bf.(NWhere).Op, bf.(NWhere).Value)
		case string:
			value, err := base.fieldValue(bf.(string))
			if err != nil {
				return result, err
			}
			repo.Where(af, value)
		}
	}
	result, err = repo.FetchKey(m.PK())

	return
}

func (base *Base) fieldValue(field string) (value interface{}, err error) {
	value, err = base.mapper.colValue(base.mapper.model, field)

	return
}

func (base *Base) Repo() *Repo {
	if base.repo != nil {
		return base.repo
	}
	base.repo = NewRepo(base.mapper.model, GetModifier(base.DBSelector()))

	return base.repo
}

func (base *Base) One(name string) (one interface{}, err error) {
	if v, ok := base.onesValue[name]; ok {
		one = v
		return
	}
	if base.onesValue[name], err = base.findOne(name); err != nil {
		return
	}
	one = base.onesValue[name]
	return
}

func (base *Base) MustOne(name string) interface{} {
	if one, err := base.One(name); err != nil {
		panic(err)
	} else {
		return one
	}
}

func (base *Base) MustMany(name string) interface{} {
	if many, err := base.Many(name); err != nil {
		panic(err)
	} else {
		return many
	}
}

func (base *Base) Many(name string) (many interface{}, err error) {
	if v, ok := base.manysValue[name]; ok {
		many = v
		return
	}
	if base.manysValue[name], err = base.findMany(name); err != nil {
		return
	}
	many = base.manysValue[name]
	return
}

func (base *Base) MarshalJSON() ([]byte, error) {
	return json.Marshal(base.Map())
}

func (base *Base) Map() map[string]interface{} {
	result := make(map[string]interface{})
	mapper := base.Mapper()
	values := mapper.modelValue(mapper.model)
	mapper.each(func(fd *fieldDescriptor) bool {
		if !fd.protected {
			result[fd.colname] = values.FieldByName(fd.fieldname).Interface()
		}
		return true
	})

	return result
}

func (base *Base) Create() error {
	return base.Repo().Create(base.mapper.model)
}

func (base *Base) Update() error {
	return base.Repo().Update(base.mapper.model)
}

func (base *Base) Delete() error {
	return base.Repo().Delete(base.mapper.model)
}

func (base *Base) Save() error {
	if base.fresh {
		return base.Create()
	}
	return base.Update()
}

func (base *Base) Fill(data map[string]interface{}) {
	for colname, val := range data {
		if val == nil {
			continue
		} else if fd, ok := base.mapper.fd(colname); !ok {
			continue
		} else if fd.protected {
			continue
		} else {
			field := base.mapper.value.FieldByName(fd.fieldname)
			field.Set(reflect.ValueOf(val))
		}
	}
}

func (base *Base) Set(colname string, val interface{}) error {
	if fd, ok := base.mapper.fd(colname); ok {
		field := base.mapper.value.FieldByName(fd.fieldname)
		field.Set(reflect.ValueOf(val))

		return nil
	}

	return errors.New("col " + colname + " not defined on model")
}

func (base *Base) Has(colname string) bool {
	return base.mapper.has(colname)
}

func (base *Base) Get(colname string) interface{} {
	if val, err := base.fieldValue(colname); err != nil {
		return nil
	} else {
		return val
	}
}

func (base *Base) PK() string {
	return base.mapper.pk
}

func New(m interface{}) interface{} {
	value := reflect.ValueOf(m)
	field := value.Elem().FieldByName("Base")
	field.Set(reflect.ValueOf(newBase(m)))
	if p, ok := m.(Preparable); ok {
		p.Prepare()
	}
	return m
}
