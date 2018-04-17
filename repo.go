package model

import (
	"context"
	"database/sql"
	"errors"
	. "github.com/yang-zzhong/go-querybuilder"
	"reflect"
	"strings"
)

type Repo struct {
	model interface{}
	conn  *sql.DB
	mm    *ModelMapper
	*Builder
}

var conn *sql.DB

func NewRepo(m interface{}, conn *sql.DB, p Modifier) *Repo {
	repo := &Repo{m, conn, NewModelMapper(m), NewBuilder(p)}
	repo.From(repo.model.(Model).TableName())

	return repo
}

func (repo *Repo) One() interface{} {
	result, _ := repo.Fetch()
	if len(result) > 0 {
		return result[0]
	}
	return nil
}

func (repo *Repo) Find(val interface{}) interface{} {
	repo.Where(repo.model.(Model).IdKey(), val.(string)).Limit(1)
	result, _ := repo.Fetch()
	if len(result) > 0 {
		return result[0]
	}
	return nil
}

func (repo *Repo) Fetch() (result []interface{}, err error) {
	result = []interface{}{}
	rows, qerr := repo.conn.Query(repo.ForQuery(), repo.Params()...)
	if qerr != nil {
		err = qerr
		return
	}
	for rows.Next() {
		columns, cerr := rows.Columns()
		if cerr != nil {
			err = cerr
			return
		}
		rerr := rows.Scan(repo.mm.ValueReceivers(columns)...)
		if rerr != nil {
			err = rerr
			return
		}
		model := reflect.ValueOf(repo.mm.Model()).Elem().Interface()
		result = append(result, model)
	}

	return
}

func (repo *Repo) UpdateRaw(data map[string]interface{}) {
	repo.conn.Exec(repo.ForUpdate(data), repo.Params()...)
}

func (repo *Repo) RemoveRaw() {
	repo.conn.Exec(repo.ForRemove(), repo.Params()...)
}

func (repo *Repo) Update(model interface{}) error {
	if err := repo.Validate(model); err != nil {
		return err
	}
	field := repo.model.(Model).IdKey()
	priValue, _ := repo.mm.FindFieldValue(model, field)
	repo.Where(field, priValue)
	data, _ := repo.mm.Extract(model)
	repo.conn.Exec(repo.ForUpdate(data), repo.Params()...)

	return nil
}

func (repo *Repo) Remove() {
	repo.conn.Exec(repo.ForRemove(), repo.Params()...)
}

func (repo *Repo) Create(model interface{}) error {
	if err := repo.Validate(model); err != nil {
		return err
	}
	row, _ := repo.mm.Extract(model)
	data := []map[string]interface{}{row}
	repo.conn.Exec(repo.ForInsert(data), repo.Params()...)
	return nil
}

func (repo *Repo) Count() int {
	rows, _ := repo.conn.Query(repo.ForCount(), repo.Params()...)
	result := 0
	for rows.Next() {
		rows.Scan(&result)
		return result
	}
	return result
}

func (repo *Repo) CreateTable() error {
	sql := "CREATE TABLE " + repo.QuotedTableName()
	rowsInfo := []string{}
	indexes := []string{}
	for _, item := range repo.mm.Fds {
		rowInfo := []string{item.FieldName, item.FieldType}
		if item.IsPk {
			rowInfo = append(rowInfo, "PRIMARY KEY")
		}
		if !item.Nullable {
			rowInfo = append(rowInfo, "NOT NULL")
		}
		if item.IsUk {
			indexSql := "CREATE UNIQUE INDEX ui_" +
				repo.model.(Model).TableName() + "_" + item.FieldName +
				" ON " + repo.model.(Model).TableName() + " (" + item.FieldName + ")"
			indexes = append(indexes, indexSql)
		}
		rowsInfo = append(rowsInfo, strings.Join(rowInfo, " "))
	}
	sql += "(\n\t" + strings.Join(rowsInfo, ",\n\t") + "\n)"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tx, err := repo.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	_, err := tx.Exec(sql)
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, index := range indexes {
		_, err := tx.Exec(index)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (repo *Repo) BeginTx(ctx context.Context, opts *sql.TxOptions) (tx *sql.Tx, err error) {
	tx, err = repo.conn.BeginTx(ctx, opts)
	return
}

func (repo *Repo) Validate(model interface{}) error {
	mValue, _ := repo.mm.modelValue(model)
	for _, item := range repo.mm.Fds {
		value := mValue.(reflect.Value).FieldByName(item.Name).Interface()
		if item.Nullable && isNull(value) {
			return errors.New(item.Name + " Not Allow Null")
		}
		if item.IsPk || item.IsUk {
			cRepo := &(*repo)
			idKey := repo.model.(Model).IdKey()
			idValue, _ := repo.mm.FindFieldValue(model, idKey)
			cRepo.Where(idKey, NEQ, idValue).Where(item.FieldName, value)
			if cRepo.Count() > 0 {
				return errors.New(item.Name + " Exists In DB")
			}
		}
	}

	return nil
}

func isNull(value interface{}) bool {
	if value == nil {
		return true
	}
	mValue := reflect.ValueOf(value)
	for mValue.Kind() == reflect.Ptr {
		mValue = mValue.Elem()
	}
	switch mValue.Kind() {
	case reflect.String:
		return mValue.Interface().(string) == ""
	}

	return false
}
