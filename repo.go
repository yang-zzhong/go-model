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
type QueryCallback func(*sql.Rows)

const (
	ONE = iota
	MANY
)

type with struct {
	name string
	t    int
}

type Repo struct {
	model    interface{}
	conn     *sql.DB
	mm       *ModelMapper
	onCreate onModify
	onUpdate onModify
	tx       *sql.Tx
	with     []with
	*Builder
}

var conn *sql.DB

func NewCustomRepo(m interface{}, conn *sql.DB, p Modifier) *Repo {
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

func NewRepo(m interface{}) (repo *Repo, err error) {
	if !Inited() {
		err = errors.New("initiate required")
		return
	}

	repo = NewCustomRepo(m, conn.db, conn.modifier)
	return
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
	for _, one := range result {
		return one
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
	for _, one := range result {
		return one
	}
	return nil
}

func (repo *Repo) QueryCallback(call QueryCallback) error {
	rows, qerr := repo.conn.Query(repo.ForQuery(), repo.Params()...)
	repo.executed()
	if qerr != nil {
		return qerr
	}
	for rows.Next() {
		call(rows)
	}
	return nil
}

func (repo *Repo) Query(key string) (result map[interface{}]interface{}, err error) {
	result = make(map[interface{}]interface{})
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
		receivers := repo.mm.ValueReceivers(columns)
		rerr := rows.Scan(receivers...)
		if rerr != nil {
			err = rerr
			return
		}
		item := repo.packItem(columns, receivers)
		result[item[key]] = item
	}

	return
}

func (repo *Repo) Fetch() (result map[string]interface{}, err error) {
	result = make(map[string]interface{})
	rows, qerr := repo.conn.Query(repo.ForQuery(), repo.Params()...)
	repo.executed()
	if qerr != nil {
		err = qerr
		return
	}
	columnValues := repo.columnValues()
	for rows.Next() {
		columns, cerr := rows.Columns()
		if cerr != nil {
			err = cerr
			return
		}
		receivers := repo.mm.ValueReceivers(columns)
		rerr := rows.Scan(receivers...)
		if rerr != nil {
			err = rerr
			return
		}
		m, id := repo.mm.Pack(columns, receivers)
		if err = repo.voc(m, columnValues); err != nil {
			return
		}
		result[id] = m
	}
	var ones []map[string]interface{}
	var manys []map[string][]interface{}
	if ones, err = repo.ones(columnValues); err != nil {
		return
	}
	if manys, err = repo.manys(columnValues); err != nil {
		return
	}
	for _, rel := range ones {
		result[rel["id"]].addOne(rel["name"], rel["model"])
	}
	for _, rel := range manys {
		result[rel["id"]].addMany(rel["name"], rel["model"])
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
	repo.onUpdate(model)
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
	var err error
	if err = repo.ValidateNullable(model); err != nil {
		return err
	}
	repo.onCreate(model)
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
	sqlang, indexes := repo.forCreateTable()
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

func (repo *Repo) forCreateTable() (sqlang string, indexes []string) {
	sqlang = "CREATE TABLE " + repo.QuotedTableName()
	indexes = []string{}
	rowsInfo := []string{}
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
	return
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
	mValue, err := repo.mm.modelValue(model)
	if err != nil {
		return err
	}
	for _, item := range repo.mm.Fds {
		value := mValue.(reflect.Value).FieldByName(item.Name).Interface()
		if !item.Nullable && isNull(value) {
			return errors.New(item.Name + " not nullable")
		}
	}

	return nil
}

func (repo *Repo) With(name string) *Repo {
	if _, _, err := repo.m.(Model).HasOne(name); err == nil {
		repo.with = append(repo.with, with{name, ONE})
	}
	if _, _, err := repo.m.(Model).HasMany(name); err == nil {
		repo.with = append(repo.with, with{name, MANY})
	}

	return repo
}

func (repo *Repo) packItem(cols []string, receivers []interface{}) map[string]interface{} {
	item := make(map[string]interface{})
	for i, col := range cols {
		item[col] = reflect.ValueOf(receivers[i]).Elem().Interface()
	}

	return item
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
