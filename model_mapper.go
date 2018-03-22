package model

type ModelMapper struct {
	Fresh  bool
	model  interface{}
	fds    []FieldDescriptor
	fdsMap map[string]int
}

type FieldDescriptor struct {
	Name      string
	FieldName string
	Value     interface{}
	Nullable  bool
	PK        bool
	Index     bool
	Uniques   []string
	FK        []string
}

func NewMM(model interface{}) *ModelMapper {
	mm = new(ModelMapper)
	mm.model = model
	mm.fd = []FieldDescriptor{}
	values := reflect.ValueOf(mm.model).Elem()
	types := reflect.TypeOf(mm.model).Elem()
	length := types.NumField()
	mm.fds = make([]FieldDescriptor, length)
	mm.fdsMap = make(map[string]int)
	for i := 0; i < length; i++ {
		fd = new(FieldDescriptor)
		fd.Name = types.Field(i).Name()
		fd.Value = values[i].Addr().Interface()
		parseTag(types.Field(i).Tag, fd)
		mm.fds[i] = *fd
		mm.fdsMap[fd.Name] = i
	}

	return mm
}

/**
 * type User struct {
 *	  Id   		int 	`db:"id,pk"`
 *	  Name 		string 	`db:"name,index"`
 *	  Age  		int		`db:"age,nil"`
 *	  Addr 		string	`db:"address,nil"`
 *	  Code 		string	`db:"code,uk"`
 *	  Area		string	`db:"area" uk:"area-area_code"`
 *	  AreaCode  string  `db:"area_code" uk:"area-area_code"`
 * }
 *
 * type Book struct {
 *	  Id		int 	`db:"id,pk"`
 *	  Title		string	`db:"title,index"`
 *	  AuthorId	int		`db:"author_id,index" fk:"user"`
 * }
 */
func parseTag(tag reflect.TagStruct, fd *FieldDescriptor) {
	db := tag.Get("db")
	space := regexp.MustCompile("\\S+")
	opts := helper.Explode(space.ReplaceAll(db, ""), ",")
	fd.FieldName = opts[0]
	fd.Nullable = helper.InArray(opts, "nil")
	fd.PK = helper.InArray(opts, "pk")
	fd.Index = helper.InArray(opts, "index")
	fd.Uniques = []string{}
	if helper.InArray(opts, "uk") {
		fd.Uniques = append(fd.Uniques, fd.FieldName)
	}
	if uk, ok := tag.Lookup("uk"); ok {
		space := regexp.MustCompile("\\S+")
		uks := helper.Explode(space.ReplaceAll(db, ""), ",")
		fd.Uniques = helper.Merge(fd.Uniques, uks)
	}
	if fk, ok := tag.Lookup("fk"); ok {
		space := regexp.MustCompile("\\S+")
		fd.FK = helper.Explode(space.ReplaceAll(db, ""), ",")
	}
}

func (mm *ModelMapper) FieldReceivers() []interface{} {
	receivers := make([]interface{}, len(mm.fds))
	for i, fd := range mm.fds {
		receivers[i] = &fd.value
	}

	return receivers
	// value := reflect.ValueOf(mm.model).Elem()
	// length := value.NumField()
	// pointers := make([]interface{}, length)
	// for i := 0; i < length; i++ {
	// 	pointers[i] = value.Field(i).Addr().Interface()
	// }

	// return pointers
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
		if len(fd.UK) == 0 {
			continue
		}
		for _, uk := range fd.UK {
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
		if len(fd.UK) == 0 {
			continue
		}
		for _, uk := range fd.UK {
			if _, ok := result[uk]; !ok {
				result[uk] = []string{fd.FieldName}
				continue
			}
			result[uk] = append(result[uk], fd.FieldName)
		}
	}

	return result
}

func (mm *ModelMapper) TableName() string {
	return mm.model.(TableNamer).TableName()
}

func (mm *ModelMapper) Describe() []FieldDescriptor {
	return mm.fds
}

func (mm *ModelMapper) Model() interface{} {
	values := reflect.ValueOf(mm.model).Elem()
	for i := 0; i < values.NumField(); i++ {
		v := values.Field(i)
		v.Set(reflect.ValueOf(mm.fds[i].value))
	}
	return mm.model
	// return reflect.ValueOf(mm.model).Elem().Interface()
}
