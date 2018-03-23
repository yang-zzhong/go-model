package model

import (
	"errors"
	helper "github.com/yang-zzhong/go-helpers"
	"regexp"
)

type TagParser struct{}

type dbResult struct {
	FieldName string
	FieldType string
	IsUk      bool
	IsPk      bool
	IsIndex   bool
	Nullable  bool
}

func (tp *TagParser) ParseDB(db string) (result *dbResult, err error) {
	result = new(dbResult)
	result.IsUk = false
	result.IsPk = false
	result.IsIndex = false
	result.Nullable = false
	opts := regexp.MustCompile("\\[.*\\]")
	rest := opts.ReplaceAllFunc(([]byte)(db), func(matched []byte) []byte {
		space := regexp.MustCompile("\\s+")
		sm := (string)(space.ReplaceAll(matched, ([]byte)("")))
		opt := helper.Explode(sm[1:len(sm)-1], ",")
		result.IsUk = helper.InStrArray(opt, "uk")
		result.IsPk = helper.InStrArray(opt, "pk")
		result.IsIndex = helper.InStrArray(opt, "index")
		result.Nullable = helper.InStrArray(opt, "nil")
		return ([]byte)("")
	})
	main := helper.Explode((string)(rest), " ")
	if len(main) < 1 {
		err = errors.New("Field Name And Type Not Assigned")
		return
	}
	if len(main) < 2 {
		err = errors.New("DB Field Type Not Assigned For \"" + main[0] + "\"")
		return
	}
	result.FieldName = main[0]
	result.FieldType = main[1]

	return
}

func (tp *TagParser) ParseUk(uk string) []string {
	space := regexp.MustCompile("\\s+")

	return helper.Explode((string)(space.ReplaceAll(([]byte)(uk), ([]byte)(""))), ",")
}

func (tp *TagParser) ParseFk(fk string) []string {
	space := regexp.MustCompile("\\s+")
	return helper.Explode((string)(space.ReplaceAll(([]byte)(fk), ([]byte)(""))), ",")
}
