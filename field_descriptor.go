package model

import (
	"strings"
)

//
// fieldDescriptor hold the col info and relate to struct field info
//
type fieldDescriptor struct {
	fieldname string
	colname   string
	coltype   string
	protected bool
	nullable  bool
	isuk      bool
	ispk      bool
	isindex   bool
}

//
// new a fieldDescriptor
//
func newFd(fieldname string, src string) *fieldDescriptor {
	fd := new(fieldDescriptor)
	fd.fieldname = fieldname
	fd.parse(src)

	return fd
}

//
// type User struct {
//    Id   		int 	`db:"id int pk,index"`
//    Name 		string 	`db:"name varchar(63) index"`
//    Age  		int		`db:"age int nil"`
//    Addr 		string	`db:"address varchar(256) nil"`
//    Code 		string	`db:"code varchar(32)"`
//    Area		string	`db:"area varchar(32)"`
//    AreaCode  string  `db:"area_code varchar(32)"`
// }
//
// type Book struct {
//    Id		int 	`db:"id int pk"`
//    Title		string	`db:"title varchar(256) index"`
//    AuthorId	int		`db:"author_id int index"`
// }
//
func (fd *fieldDescriptor) parse(src string) {
	arr := strings.Split(src, "|")
	fd.colname = strings.Trim(arr[0], " ")
	fd.coltype = strings.Trim(arr[1], " ")
	if len(arr) == 3 {
		opt := strings.Split(arr[2], ",")
		for _, o := range opt {
			switch strings.Trim(o, " ") {
			case "pk":
				fd.ispk = true
			case "nil":
				fd.nullable = true
			case "index":
				fd.isindex = true
			case "protected":
				fd.protected = true
			}
		}
	}
}
