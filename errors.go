package model

import (
	helpers "github.com/yang-zzhong/go-helpers"
)

const (
	ERR_SQL = iota
	ERR_SQL_CONN
	ERR_SCAN
	ERR_NEXUS_UNDEFINED
	ERR_DATA_NOT_FOUND
	ERR_COL_UNDEFINED
	ERR_UNKNOWN_COLTYPE
)

type Error struct {
	Code int
	err  error
}

func (err *Error) Error() string {
	return err.err.Error()
}

func IsModelErr(err error) bool {
	return helpers.InstanceOf(err, &Error{})
}
