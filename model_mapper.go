package model

import (
	"errors"
	helper "github.com/yang-zzhong/go-helpers"
	"reflect"
	"strings"
)

type ModelMapper struct {
	Fresh bool
	model interface{}
	Fds   map[string]*FieldDescriptor
	FnFds map[string]*FieldDescriptor
}

type FieldDescriptor struct {
	Name      string
	FieldName string
	FieldType string
	Nullable  bool
	IsUk      bool
	IsPk      bool
	IsIndex   bool
}

func NewModelMapper(model interface{}) *ModelMapper {
	mm := new(ModelMapper)
	mm.model = model
	types := reflect.TypeOf(mm.model).Elem()
	length := types.NumField()
	mm.Fds = make(map[string]*FieldDescriptor)
	mm.FnFds = make(map[string]*FieldDescriptor)
	for i := 0; i < length; i++ {
		field := types.Field(i)
		fd := new(FieldDescriptor)
		fd.Name = field.Name
		parseTag(field.Tag, fd)
		mm.Fds[fd.Name] = fd
		mm.FnFds[fd.FieldName] = fd
	}

	return mm
}

/**
 * type User struct {
 *	  Id   		int 	`db:"id int pk,index"`
 *	  Name 		string 	`db:"name varchar(63) index"`
 *	  Age  		int		`db:"age int nil"`
 *	  Addr 		string	`db:"address varchar(256) nil"`
 *	  Code 		string	`db:"code varchar(32)"`
 *	  Area		string	`db:"area varchar(32)"`
 *	  AreaCode  string  `db:"area_code varchar(32)"`
 * }
 *
 * type Book struct {
 *	  Id		int 	`db:"id int pk"`
 *	  Title		string	`db:"title varchar(256) index"`
 *	  AuthorId	int		`db:"author_id int index"`
 * }
 */
func parseTag(tag reflect.StructTag, fd *FieldDescriptor) {
	dbArray := strings.Split(tag.Get("db"), " ")
	fd.FieldName = dbArray[0]
	fd.FieldType = dbArray[1]
	if len(dbArray) == 3 {
		opt := strings.Split(dbArray[2], ",")
		fd.IsPk = helper.InStrArray(opt, "pk")
		fd.IsUk = helper.InStrArray(opt, "uk")
		fd.IsIndex = helper.InStrArray(opt, "index")
		fd.Nullable = helper.InStrArray(opt, "nil")
	}
}

func (mm *ModelMapper) ValueReceivers(columns []string) []interface{} {
	value := reflect.ValueOf(mm.model).Elem()
	pointers := make([]interface{}, len(columns))
	for i, fieldName := range columns {
		name := mm.FnFds[fieldName].Name
		pointers[i] = value.FieldByName(name).Addr().Interface()
	}

	return pointers
}

func (mm *ModelMapper) Model() interface{} {
	return mm.model
}

func (mm *ModelMapper) Extract(model interface{}) (result map[string]interface{}, err error) {
	result = make(map[string]interface{})
	mValue, err := mm.modelValue(model)
	for _, item := range mm.Fds {
		result[item.FieldName] = mValue.(reflect.Value).FieldByName(item.Name).Interface()
	}

	return
}

func (mm *ModelMapper) DbFieldValue(model interface{}, field string) (result interface{}, err error) {
	mValue, perr := mm.modelValue(model)
	if perr != nil {
		err = perr
		return
	}
	if desc, ok := mm.FnFds[field]; ok {
		result = mValue.(reflect.Value).FieldByName(desc.Name).Interface()
	}
	return
}

func (mm *ModelMapper) FieldValue(model interface{}, field string) (result interface{}, err error) {
	mValue, perr := mm.modelValue(model)
	if perr != nil {
		err = perr
		return
	}
	if desc, ok := mm.Fds[field]; ok {
		result = mValue.(reflect.Value).FieldByName(desc.Name).Interface()
	}
	return
}

func (mm *ModelMapper) modelValue(model interface{}) (result interface{}, err error) {
	mType := reflect.TypeOf(model)
	mmType := reflect.TypeOf(mm.model)
	if mType.Name() != mmType.Name() {
		err = errors.New("model type error")
		return
	}
	mValue := reflect.ValueOf(model)
	for mValue.Kind() == reflect.Ptr {
		mValue = mValue.Elem()
	}

	result = mValue
	return
}
