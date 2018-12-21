package model

import (
	"context"
	"database/sql"
	. "github.com/yang-zzhong/go-querybuilder"
)

const DEFAULT = "default"

var (
	instances *Db
	dbpairs   map[string]dbpair
)

type DB interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
}

// tx callback type
type txhandler func(t *sql.Tx) error

type dbpair struct {
	modifier Modifier
	db       *Db
}

type Db struct {
	txs []*sql.Tx
	*sql.DB
}

func init() {
	dbpairs = make(map[string]dbpair)
}

func RegisterDefaultDB(db *sql.DB, m Modifier) {
	RegisterDB(db, m, DEFAULT)
}

func RegisterDB(db *sql.DB, m Modifier, name string) {
	dbpairs[name] = dbpair{m, &Db{[]*sql.Tx{}, db}}
}

func GetDB(name string) *Db {
	if p, ok := dbpairs[name]; ok {
		return p.db
	}
	panic("db config '" + name + "' not found")
}

func GetDefaultDB() *Db {
	return GetDB(DEFAULT)
}

func GetModifier(name string) Modifier {
	if p, ok := dbpairs[name]; ok {
		return p.modifier
	}

	panic("db config '" + name + "' not found")
}

func GetDefaultModifier() Modifier {
	return GetModifier(DEFAULT)
}

func UnregisterDefaultDB() {
	UnregisterDB(DEFAULT)
}

func UnregisterDB(name string) {
	delete(dbpairs, name)
}

func (db *Db) TxContext(handle txhandler, ctx context.Context, opts *sql.TxOptions) error {
	var err error
	db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	defer func() {
		if e := recover(); e != nil {
			db.Rollback()
			panic(e)
		}
	}()
	if err := handle(db.tx()); err != nil {
		db.Rollback()
		return err
	}
	if err := db.Commit(); err != nil {
		db.Rollback()
		return err
	}

	return nil
}

func (db *Db) Tx(handle txhandler) error {
	var err error
	db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if e := recover(); e != nil {
			db.Rollback()
			panic(e)
		}
	}()
	if err := handle(db.tx()); err != nil {
		db.Rollback()
		return err
	}
	if err := db.Commit(); err != nil {
		db.Rollback()
		return err
	}

	return nil
}

func (db *Db) Begin() error {
	tx, err := db.DB.Begin()
	db.txs = append(db.txs, tx)

	return err
}

func (db *Db) BeginTx(ctx context.Context, opts *sql.TxOptions) error {
	tx, err := db.DB.BeginTx(ctx, opts)
	db.txs = append(db.txs, tx)

	return err
}

func (db *Db) Commit() error {
	tx := db.tx()
	if tx == nil {
		panic("please begin before commit")
	}
	err := tx.Commit()
	db.poptx()
	return err
}

func (db *Db) Rollback() error {
	tx := db.tx()
	if tx == nil {
		panic("please begin before rollback")
	}
	err := tx.Rollback()
	db.poptx()
	return err
}

func (db *Db) Exec(query string, args ...interface{}) (sql.Result, error) {
	tx := db.tx()
	if tx != nil {
		return tx.Exec(query, args...)
	}

	return db.DB.Exec(query, args...)
}

func (db *Db) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	tx := db.tx()
	if tx != nil {
		return tx.ExecContext(ctx, query, args...)
	}

	return db.DB.ExecContext(ctx, query, args...)
}

func (db *Db) Prepare(query string) (*sql.Stmt, error) {
	tx := db.tx()
	if tx != nil {
		return tx.Prepare(query)
	}

	return db.DB.Prepare(query)
}

func (db *Db) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	tx := db.tx()
	if tx != nil {
		return tx.PrepareContext(ctx, query)
	}

	return db.DB.PrepareContext(ctx, query)
}

func (db *Db) Query(query string, args ...interface{}) (*sql.Rows, error) {
	tx := db.tx()
	if tx != nil {
		return tx.Query(query, args...)
	}

	return db.DB.Query(query, args...)
}

func (db *Db) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	tx := db.tx()
	if tx != nil {
		return tx.QueryContext(ctx, query, args...)
	}

	return db.DB.QueryContext(ctx, query, args...)
}

func (db *Db) QueryRow(query string, args ...interface{}) *sql.Row {
	tx := db.tx()
	if tx != nil {
		return tx.QueryRow(query, args...)
	}

	return db.DB.QueryRow(query, args...)
}

func (db *Db) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	tx := db.tx()
	if tx != nil {
		return tx.QueryRowContext(ctx, query, args...)
	}

	return db.DB.QueryRowContext(ctx, query, args...)
}

func (db *Db) tx() *sql.Tx {
	if len(db.txs) == 0 {
		return nil
	}
	return db.txs[len(db.txs)-1]
}

func (db *Db) poptx() {
	if len(db.txs) == 1 {
		db.txs = []*sql.Tx{}
		return
	}
	db.txs = db.txs[:len(db.txs)-2]
}
