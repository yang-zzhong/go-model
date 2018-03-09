package model

import (
	helper "github.com/yang-zzhong/go-helpers"
	"reflect"
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
	PK        bool
	Index     bool
	Uniques   []string
	FK        []string
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
 *	  Id   		int 	`db:"id int[pk]"`
 *	  Name 		string 	`db:"name varchar(63)[index]"`
 *	  Age  		int		`db:"age int[nil]"`
 *	  Addr 		string	`db:"address varchar(256)[nil]"`
 *	  Code 		string	`db:"code varchar(32)[uk]"`
 *	  Area		string	`db:"area varchar(32)" uk:"area-area_code"`
 *	  AreaCode  string  `db:"area_code varchar(32)" uk:"area-area_code"`
 * }
 *
 * type Book struct {
 *	  Id		int 	`db:"id int[pk]"`
 *	  Title		string	`db:"title varchar(256)[index]"`
 *	  AuthorId	int		`db:"author_id int[index]"`
 * }
 */
func parseTag(tag reflect.StructTag, fd *FieldDescriptor) {
	parser := new(TagParser)
	dbr, _ := parser.ParseDB(tag.Get("db"))
	fd.FieldName = dbr.FieldName
	fd.FieldType = dbr.FieldType
	fd.PK = dbr.IsPk
	fd.Index = dbr.IsIndex
	fd.Nullable = dbr.Nullable
	fd.Uniques = []string{}
	fd.FK = []string{}
	if dbr.IsUk {
		fd.Uniques = append(fd.Uniques, fd.FieldName)
	}
	if uk, ok := tag.Lookup("uk"); ok {
		fd.Uniques = helper.MergeStrArray(fd.Uniques, parser.ParseUk(uk))
	}
	if fk, ok := tag.Lookup("fk"); ok {
		fd.FK = parser.ParseFk(fk)
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

func (mm *ModelMapper) IndexFields() []string {
	result := []string{}
	for _, fd := range mm.fds {
		if fd.Index {
			result = append(result, fd.FieldName)
		}
	}

	return result
}

func (mm *ModelMapper) PK() []string {
	result := []string{}
	for _, fd := range mm.fds {
		if fd.PK {
			result = append(result, fd.FieldName)
		}
	}

	return result
}

func (mm *ModelMapper) UK() [][]string {
	result := [][]string{}
	temp := make(map[string][]string)
	for _, fd := range mm.fds {
		if len(fd.Uniques) == 0 {
			continue
		}
		for _, uk := range fd.Uniques {
			if _, ok := temp[uk]; !ok {
				temp[uk] = []string{fd.FieldName}
				continue
			}
			temp[uk] = append(temp[uk], fd.FieldName)
		}
	}
	for _, fields := range temp {
		result = append(result, fields)
	}

	return result
}

func (mm *ModelMapper) FK() map[string][]string {
	result := make(map[string][]string)
	for _, fd := range mm.fds {
		if len(fd.FK) == 0 {
			continue
		}
		for _, fk := range fd.FK {
			if _, ok := result[fk]; !ok {
				result[fk] = []string{fd.FieldName}
				continue
			}
			result[fk] = append(result[fk], fd.FieldName)
		}
	}

	return result
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
