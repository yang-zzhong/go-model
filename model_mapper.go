package model

import (
	"database/sql"
	"errors"
	"reflect"
	"time"
)

type fdhandler func(fd *fieldDescriptor) bool

type ModelMapper struct {
	model     interface{}
	value     reflect.Value
	fds       map[string]*fieldDescriptor
	field2col map[string]string
}

func NewModelMapper(model interface{}) *ModelMapper {
	mm := new(ModelMapper)
	mm.model = model
	mm.fds = make(map[string]*fieldDescriptor)
	mm.value = mm.modelValue(mm.model)
	mm.field2col = make(map[string]string)
	types := reflect.TypeOf(mm.model).Elem()
	length := types.NumField()
	for i := 0; i < length; i++ {
		field := types.Field(i)
		td := field.Tag.Get("db")
		if td == "" {
			continue
		}
		fd := newFd(field.Name, td)
		mm.field2col[fd.fieldname] = fd.colname
		mm.fds[fd.colname] = fd
	}

	return mm
}

func (mm *ModelMapper) each(handle fdhandler) {
	for _, fd := range mm.fds {
		if !handle(fd) {
			break
		}
	}
}

func (mm *ModelMapper) fd(colname string) (fd *fieldDescriptor, ok bool) {
	fd, ok = mm.fds[colname]
	return
}

func (mm *ModelMapper) cols(columns []string) (result []interface{}, err error) {
	pointers := make([]interface{}, len(columns))
	var fd *fieldDescriptor
	var ok bool
	var field interface{}
	for i, colname := range columns {
		if fd, ok = mm.fd(colname); !ok {
			var value int64
			pointers[i] = &value
			continue
		}
		field = mm.value.FieldByName(fd.fieldname).Interface()
		if converter, ok := mm.model.(ValueConverter); ok {
			field = converter.DBValue(colname, field)
		}
		if fd.nullable {
			switch field.(type) {
			case string:
				pointers[i] = new(sql.NullString)
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				pointers[i] = new(sql.NullInt64)
			case float32, float64:
				pointers[i] = new(sql.NullFloat64)
			case time.Time:
				pointers[i] = new(NullTime)
			case bool:
				pointers[i] = new(sql.NullBool)
			case sql.NullString:
				pointers[i] = new(sql.NullString)
			case sql.NullBool:
				pointers[i] = new(sql.NullBool)
			case sql.NullFloat64:
				pointers[i] = new(sql.NullBool)
			case sql.NullInt64:
				pointers[i] = new(sql.NullInt64)
			case NullTime:
				pointers[i] = new(NullTime)
			default:
				err = errors.New("unknown type of field " + colname)
			}
		} else {
			switch field.(type) {
			case string:
				var value string
				pointers[i] = &value
			case int:
				var value int
				pointers[i] = &value
			case int8:
				var value int8
				pointers[i] = &value
			case int16:
				var value int16
				pointers[i] = &value
			case int32:
				var value int32
				pointers[i] = &value
			case int64:
				var value int64
				pointers[i] = &value
			case uint:
				var value uint
				pointers[i] = &value
			case uint8:
				var value uint8
				pointers[i] = &value
			case uint16:
				var value uint16
				pointers[i] = &value
			case uint32:
				var value uint32
				pointers[i] = &value
			case uint64:
				var value uint64
				pointers[i] = &value
			case float32:
				var value float32
				pointers[i] = &value
			case float64:
				var value float64
				pointers[i] = &value
			case time.Time:
				pointers[i] = new(time.Time)
			case bool:
				var value bool
				pointers[i] = &value
			default:
				err = errors.New("unknown type of field " + colname)
			}
		}
	}
	result = pointers
	return
}

func (mm *ModelMapper) pack(columns []string, cols []interface{}, key string) (model interface{}, id interface{}, err error) {
	var fd *fieldDescriptor
	var ok bool
	var converter ValueConverter
	v := reflect.New(reflect.ValueOf(mm.model).Elem().Type()).Elem()
	for i, colname := range columns {
		if fd, ok = mm.fd(colname); !ok {
			continue
		}
		field := v.FieldByName(fd.fieldname)
		col := reflect.ValueOf(cols[i]).Elem().Interface()
		if converter, ok = mm.model.(ValueConverter); ok {
			if val, catched := converter.Value(colname, col); catched {
				field.Set(val)
				if colname == key {
					id = val.Interface()
				}
				continue
			}
		}
		var value reflect.Value
		if fd.nullable {
			switch field.Interface().(type) {
			case int:
				t := col.(sql.NullInt64)
				if t.Valid {
					value = reflect.ValueOf(int(t.Int64))
				} else {
					value = reflect.ValueOf(0)
				}
			case int8:
				t := col.(sql.NullInt64)
				if t.Valid {
					value = reflect.ValueOf(int8(t.Int64))
				} else {
					value = reflect.ValueOf(0)
				}
			case int16:
				t := col.(sql.NullInt64)
				if t.Valid {
					value = reflect.ValueOf(int16(t.Int64))
				} else {
					value = reflect.ValueOf(0)
				}
			case int32:
				t := col.(sql.NullInt64)
				if t.Valid {
					value = reflect.ValueOf(int32(t.Int64))
				} else {
					value = reflect.ValueOf(0)
				}
			case int64:
				t := col.(sql.NullInt64)
				if t.Valid {
					value = reflect.ValueOf(int64(t.Int64))
				} else {
					value = reflect.ValueOf(0)
				}
			case uint:
				t := col.(sql.NullInt64)
				if t.Valid {
					value = reflect.ValueOf(int64(t.Int64))
				} else {
					value = reflect.ValueOf(0)
				}
			case uint8:
				t := col.(sql.NullInt64)
				if t.Valid {
					value = reflect.ValueOf(uint8(t.Int64))
				} else {
					value = reflect.ValueOf(0)
				}
			case uint16:
				t := col.(sql.NullInt64)
				if t.Valid {
					value = reflect.ValueOf(uint16(t.Int64))
				} else {
					value = reflect.ValueOf(0)
				}
			case uint32:
				t := col.(sql.NullInt64)
				if t.Valid {
					value = reflect.ValueOf(uint32(t.Int64))
				} else {
					value = reflect.ValueOf(0)
				}
			case uint64:
				t := col.(sql.NullInt64)
				if t.Valid {
					value = reflect.ValueOf(uint64(t.Int64))
				} else {
					value = reflect.ValueOf(0)
				}
			case float32:
				t := col.(sql.NullFloat64)
				if t.Valid {
					value = reflect.ValueOf(float32(t.Float64))
				} else {
					value = reflect.ValueOf(0.0)
				}
			case float64:
				t := col.(sql.NullFloat64)
				if t.Valid {
					value = reflect.ValueOf(float64(t.Float64))
				} else {
					value = reflect.ValueOf(0.0)
				}
			case bool:
				t := col.(sql.NullBool)
				if t.Valid {
					value = reflect.ValueOf(t.Bool)
				} else {
					value = reflect.ValueOf(false)
				}
			case string:
				t := col.(sql.NullString)
				if t.Valid {
					value = reflect.ValueOf(t.String)
				} else {
					value = reflect.ValueOf("")
				}
			case time.Time:
				t := col.(NullTime)
				if t.Valid {
					value = reflect.ValueOf(t.Time)
				}
			default:
				err = errors.New("unknown type of " + fd.colname)
				return
			}
		} else {
			value = reflect.ValueOf(col)
		}
		if value.IsValid() {
			field.Set(value)
		}
		if colname == mm.model.(Model).PK() {
			id = value.Interface()
		}
	}
	model = v.Addr().Interface()
	if base, ok := GetBase(mm.model); ok {
		b := NewBase(model)
		b.fresh = false
		b.oncreate = base.oncreate
		b.onupdate = base.onupdate
		b.ondelete = base.ondelete
		b.ones = base.ones
		b.manys = base.manys
		SetBase(model, b)
	}
	return
}

func (mm *ModelMapper) extract(model interface{}) (result map[string]interface{}) {
	result = make(map[string]interface{})
	values := mm.modelValue(model)
	for _, fd := range mm.fds {
		var value interface{}
		value = values.FieldByName(fd.fieldname).Interface()
		if converter, ok := model.(ValueConverter); ok {
			result[fd.colname] = converter.DBValue(fd.colname, value)
			continue
		}
		switch value.(type) {
		case time.Time:
			if value.(time.Time).IsZero() {
				result[fd.colname] = nil
			} else {
				result[fd.colname] = value
			}
		default:
			result[fd.colname] = value
		}
	}

	return
}

func (mm *ModelMapper) colValue(model interface{}, colname string) (result interface{}, err error) {
	values := mm.modelValue(model)
	if fd, ok := mm.fd(colname); ok {
		result = values.FieldByName(fd.fieldname).Interface()
	} else {
		err = errors.New("field " + colname + " not defined")
	}
	return
}

func (mm *ModelMapper) fieldValue(model interface{}, fieldname string) (result interface{}, err error) {
	colname := mm.field2col[fieldname]

	return mm.colValue(model, colname)
}

func (mm *ModelMapper) modelValue(model interface{}) reflect.Value {
	value := reflect.ValueOf(model)
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	return value
}
