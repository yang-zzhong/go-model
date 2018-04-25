package model

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	helper "github.com/yang-zzhong/go-helpers"
	"reflect"
	"strings"
	"time"
)

type ModelMapper struct {
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
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			continue
		}
		fd := new(FieldDescriptor)
		fd.Name = field.Name
		parseTag(dbTag, fd)
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
func parseTag(dbTag string, fd *FieldDescriptor) {
	dbArray := strings.Split(dbTag, " ")
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
	pointers := make([]interface{}, len(columns))
	values := reflect.ValueOf(mm.model).Elem()
	for i, fieldName := range columns {
		field := values.FieldByName(mm.FnFds[fieldName].Name).Interface()
		if converter, ok := mm.model.(ValueConverter); ok {
			field = converter.DBValue(fieldName, field)
		}
		switch field.(type) {
		case string:
			value := ""
			pointers[i] = &value
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			value := 0
			pointers[i] = &value
		case float32, float64:
			value := 0.0
			pointers[i] = &value
		case sql.NullString:
			value := new(sql.NullString)
			pointers[i] = value
		case sql.NullBool:
			value := new(sql.NullBool)
			pointers[i] = value
		case sql.NullFloat64:
			value := new(sql.NullBool)
			pointers[i] = value
		case sql.NullInt64:
			value := new(sql.NullInt64)
			pointers[i] = value
		case time.Time:
			pointers[i] = new(time.Time)
		case NullTime:
			pointers[i] = new(NullTime)
		}
	}

	return pointers
}

func (mm *ModelMapper) Pack(columns []string, valueReceivers []interface{}) interface{} {
	values := reflect.ValueOf(mm.model).Elem()
	for i, fieldName := range columns {
		field := values.FieldByName(mm.FnFds[fieldName].Name)
		elem := reflect.ValueOf(valueReceivers[i])
		value := elem.Elem().Interface()
		if converter, ok := mm.model.(ValueConverter); ok {
			val, catched := converter.Value(fieldName, value)
			if catched {
				field.Set(val)
				continue
			}
		}
		switch field.Kind() {
		case reflect.Bool:
			field.Set(reflect.ValueOf(value.(bool)))
		case reflect.Int:
			field.Set(reflect.ValueOf(value.(int)))
		case reflect.Int8:
			field.Set(reflect.ValueOf(value.(int8)))
		case reflect.Int16:
			field.Set(reflect.ValueOf(value.(int16)))
		case reflect.Int32:
			field.Set(reflect.ValueOf(value.(int32)))
		case reflect.Int64:
			field.Set(reflect.ValueOf(value.(int64)))
		case reflect.Uint:
			field.Set(reflect.ValueOf(value.(uint)))
		case reflect.Uint8:
			field.Set(reflect.ValueOf(value.(uint8)))
		case reflect.Uint16:
			field.Set(reflect.ValueOf(value.(uint16)))
		case reflect.Uint32:
			field.Set(reflect.ValueOf(value.(uint32)))
		case reflect.Uint64:
			field.Set(reflect.ValueOf(value.(uint64)))
		case reflect.Float32:
			field.Set(reflect.ValueOf(value.(float32)))
		case reflect.Float64:
			field.Set(reflect.ValueOf(value.(float64)))
		case reflect.Slice:
			field.Set(reflect.ValueOf(value.([]string)))
		case reflect.String:
			field.Set(reflect.ValueOf(value.(string)))
		case reflect.Struct:
			switch value.(type) {
			case sql.NullString:
				field.Set(reflect.ValueOf(value.(sql.NullString)))
			case sql.NullBool:
				field.Set(reflect.ValueOf(value.(sql.NullBool)))
			case sql.NullFloat64:
				field.Set(reflect.ValueOf(value.(sql.NullFloat64)))
			case sql.NullInt64:
				field.Set(reflect.ValueOf(value.(sql.NullInt64)))
			case time.Time:
				field.Set(reflect.ValueOf(value.(time.Time)))
			case NullTime:
				field.Set(reflect.ValueOf(value.(NullTime)))
			}
		}
	}

	return mm.model
}

func (mm *ModelMapper) Extract(model interface{}) (result map[string]interface{}, err error) {
	result = make(map[string]interface{})
	mValue, err := mm.modelValue(model)
	if err != nil {
		return
	}
	for _, item := range mm.Fds {
		value := mValue.(reflect.Value).FieldByName(item.Name).Interface()
		if converter, ok := model.(ValueConverter); ok {
			result[item.FieldName] = converter.DBValue(item.FieldName, value)
			continue
		}
		switch value.(type) {
		case time.Time:
			if value.(time.Time).IsZero() {
				result[item.FieldName] = nil
				continue
			}
		}
		result[item.FieldName] = value
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

type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// Scan implements the Scanner interface.
func (nt *NullTime) Scan(value interface{}) error {
	nt.Time, nt.Valid = value.(time.Time)
	return nil
}

// Value implements the driver Valuer interface.
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}
