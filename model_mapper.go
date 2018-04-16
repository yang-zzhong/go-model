package model

import (
	helper "github.com/yang-zzhong/go-helpers"
	"reflect"
	"strings"
)

type ModelMapper struct {
	Fresh bool
	model interface{}
	fds   map[string]*FieldDescriptor
	fnFds map[string]*FieldDescriptor
}

type FieldDescriptor struct {
	Name      string
	FieldName string
	FieldType string
	Nullable  bool
	UK        bool
	PK        bool
	Index     bool
}

func NewModelMapper(model interface{}) *ModelMapper {
	mm := new(ModelMapper)
	mm.model = model
	types := reflect.TypeOf(mm.model).Elem()
	length := types.NumField()
	mm.fds = make(map[string]*FieldDescriptor)
	mm.fnFds = make(map[string]*FieldDescriptor)
	for i := 0; i < length; i++ {
		field := types.Field(i)
		fd := new(FieldDescriptor)
		fd.Name = field.Name
		parseTag(field.Tag, fd)
		mm.fds[fd.Name] = fd
		mm.fnFds[fd.FieldName] = fd
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
		fd.PK = helper.InStrArray(opt, "pk")
		fd.UK = helper.InStrArray(opt, "uk")
		fd.Index = helper.InStrArray(opt, "index")
		fd.Nullable = helper.InStrArray(opt, "nil")
	}
}

func (mm *ModelMapper) ValueReceivers(columns []string) []interface{} {
	value := reflect.ValueOf(mm.model).Elem()
	pointers := make([]interface{}, len(columns))
	for i, fieldName := range columns {
		name := mm.fnFds[fieldName].Name
		pointers[i] = value.FieldByName(name).Addr().Interface()
	}

	return pointers
}

func (mm *ModelMapper) TableName() string {
	return mm.model.(TableNamer).TableName()
}

func (mm *ModelMapper) Describe() map[string]*FieldDescriptor {
	return mm.fds
}

func (mm *ModelMapper) Model() interface{} {
	return mm.model
}
