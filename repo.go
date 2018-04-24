package model

import (
	"context"
	"database/sql"
	"errors"
	. "github.com/yang-zzhong/go-querybuilder"
	"reflect"
	"strings"
)

type onModify func(model interface{})
type txCall func(tx *sql.Tx) error

type Repo struct {
	model    interface{}
	conn     *sql.DB
	mm       *ModelMapper
	onCreate onModify
	onUpdate onModify
	tx       *sql.Tx
	*Builder
}

var conn *sql.DB

func NewRepo(m interface{}, conn *sql.DB, p Modifier) *Repo {
	repo := new(Repo)
	repo.model = m
	repo.conn = conn
	repo.mm = NewModelMapper(m)
	repo.onCreate = func(model interface{}) {}
	repo.onUpdate = func(model interface{}) {}
	repo.Builder = NewBuilder(p)
	repo.From(repo.model.(Model).TableName())

	return repo
}

func (repo *Repo) WithTx(tx *sql.Tx) *Repo {
	repo.tx = tx
	return repo
}

func (repo *Repo) OnCreate(oncreate onModify) *Repo {
	repo.onCreate = oncreate
	return repo
}

func (repo *Repo) OnUpdate(onupdate onModify) *Repo {
	repo.onUpdate = onupdate
	return repo
}

func (repo *Repo) CallOnUpdate(model interface{}) *Repo {
	repo.onUpdate(model)
	return repo
}

func (repo *Repo) CallOnCreate(model interface{}) *Repo {
	repo.onCreate(model)
	return repo
}

func (repo *Repo) One() interface{} {
	result, _ := repo.Fetch()
	if len(result) > 0 {
		return result[0]
	}
	return nil
}

func (repo *Repo) executed() {
	repo.tx = nil
	repo.Builder.Init()
}

func (repo *Repo) Find(val interface{}) interface{} {
	repo.Where(repo.model.(Model).PK(), val.(string)).Limit(1)
	result, _ := repo.Fetch()
	if len(result) > 0 {
		return result[0]
	}
	return nil
}

func (repo *Repo) Fetch() (result []interface{}, err error) {
	result = []interface{}{}
	rows, qerr := repo.conn.Query(repo.ForQuery(), repo.Params()...)
	repo.executed()
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
	if repo.tx != nil {
		repo.tx.Exec(repo.ForUpdate(data), repo.Params()...)
	}
	repo.conn.Exec(repo.ForUpdate(data), repo.Params()...)
	repo.executed()
}

func (repo *Repo) RemoveRaw() {
	if repo.tx != nil {
		repo.tx.Exec(repo.ForRemove(), repo.Params()...)
	}
	repo.conn.Exec(repo.ForRemove(), repo.Params()...)
	repo.executed()
}

func (repo *Repo) Update(model interface{}) error {
	repo.onUpdate(&model)
	var err error
	if err = repo.ValidateNullable(model); err != nil {
		return err
	}
	field := repo.model.(Model).PK()
	priValue, _ := repo.mm.DbFieldValue(model, field)
	repo.Where(field, priValue)
	data, _ := repo.mm.Extract(model)
	if repo.tx != nil {
		_, err = repo.tx.Exec(repo.ForUpdate(data), repo.Params()...)
	} else {
		_, err = repo.conn.Exec(repo.ForUpdate(data), repo.Params()...)
	}
	repo.executed()

	return nil
}

func (repo *Repo) Remove() error {
	var err error
	if repo.tx != nil {
		_, err = repo.tx.Exec(repo.ForRemove(), repo.Params()...)
	} else {
		_, err = repo.conn.Exec(repo.ForRemove(), repo.Params()...)
	}
	repo.executed()
	return err
}

func (repo *Repo) Create(model interface{}) error {
	repo.onCreate(&model)
	var err error
	if err = repo.ValidateNullable(model); err != nil {
		return err
	}
	row, _ := repo.mm.Extract(model)
	data := []map[string]interface{}{row}
	if repo.tx != nil {
		_, err = repo.tx.Exec(repo.ForInsert(data), repo.Params()...)
	} else {
		_, err = repo.conn.Exec(repo.ForInsert(data), repo.Params()...)
	}
	repo.executed()

	return err
}

func (repo *Repo) Count() int {
	rows, err := repo.conn.Query(repo.ForCount(), repo.Params()...)
	if err != nil {
		panic(err)
	}
	repo.Builder.Init()
	var result int
	for rows.Next() {
		rows.Scan(&result)
		break
	}
	return result
}

func (repo *Repo) CreateTable() error {
	sqlang := "CREATE TABLE " + repo.QuotedTableName()
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
	sqlang += "(\n\t" + strings.Join(rowsInfo, ",\n\t") + "\n)"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return repo.Tx(func(tx *sql.Tx) error {
		_, err := tx.Exec(sqlang)
		if err != nil {
			return err
		}
		for _, index := range indexes {
			_, err := tx.Exec(index)
			if err != nil {
				return err
			}
		}
		return nil
	}, ctx, nil)
}

func (repo *Repo) Tx(txcall txCall, ctx context.Context, opts *sql.TxOptions) error {
	tx, err := repo.conn.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	if err := txcall(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (repo *Repo) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return repo.conn.BeginTx(ctx, opts)
}

func (repo *Repo) ValidateNullable(model interface{}) error {
	mValue, _ := repo.mm.modelValue(model)
	for _, item := range repo.mm.Fds {
		value := mValue.(reflect.Value).FieldByName(item.Name).Interface()
		if !item.Nullable && isNull(value) {
			return errors.New(item.Name + " not nullable")
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
