package model

import (
	"context"
	"database/sql"
	. "github.com/yang-zzhong/go-querybuilder"
	"sync"
)

var (
	instance *Db
	once     sync.Once
	db       *sql.DB
	modifier Modifier
)

// tx callback type
type txhandler func(t *sql.Tx) error

type Db struct {
	txs []*sql.Tx
	*sql.DB
}

func Config(sdb *sql.DB, m Modifier) {
	db = sdb
	modifier = m
}

func GetDB() *Db {
	once.Do(func() {
		instance = new(Db)
		instance.DB = db
		instance.tx = []*sql.Tx{}
	})

	return instance
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
	if err := handle(tx); err != nil {
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
	if err := handle(tx); err != nil {
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
	var err error
	db.tx, err = db.DB.Begin()

	return err
}

func (db *Db) BeginTx(ctx context.Context, opts *sql.TxOptions) error {
	var err error
	tx, err = db.DB.BeginTx(ctx, opts)
	db.txs = append(db.txs, tx)

	return err
}

func (db *Db) Commit() error {
	tx = db.tx()
	if tx != nil {
		panic("please begin before commit")
	}
	err := db.tx.Commit()
	db.poptx()
	return err
}

func (db *Db) Rollback() error {
	tx = db.tx()
	if tx != nil {
		panic("please begin before rollback")
	}
	err := db.tx.Rollback()
	db.poptx()
	db.tx = nil
}

func (db *Db) Exec(query string, args ...interface{}) (sql.Result, error) {
	tx = db.tx()
	if tx != nil {
		return db.tx.Exec(query, args)
	}

	return db.DB.Exec(query, args)
}

func (db *Db) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	tx = db.tx()
	if tx != nil {
		return db.tx.ExecContext(ctx, query, args)
	}

	return db.DB.Exec(ctx, query, args)
}

func (db *Db) Prepare(query string) (*Stmt, error) {
	tx = db.tx()
	if tx != nil {
		return db.tx.Prepare(query)
	}

	return db.DB.Prepare(query)
}

func (db *Db) PrepareContext(ctx context.Context, query string) (*Stmt, error) {
	tx = db.tx()
	if tx != nil {
		return db.tx.PrepareContext(query)
	}

	return db.DB.PrepareContext(query)
}

func (db *Db) Query(query string, args ...interface{}) (*sql.Rows, error) {
	tx = db.tx()
	if tx != nil {
		return db.tx.Query(query, args)
	}

	return db.DB.Query(query, args)
}

func (db *Db) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	tx = db.tx()
	if tx != nil {
		return db.tx.QueryContext(ctx, query, args)
	}

	return db.DB.QueryContext(ctx, query, args)
}

func (db *Db) QueryRow(query string, args ...interface{}) *sql.Row {
	tx = db.tx()
	if tx != nil {
		return db.tx.QueryRow(query, args)
	}

	return db.DB.QueryRow(query, args)
}

func (db *Db) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	tx = db.tx()
	if tx != nil {
		return tx.QueryRowContext(ctx, query, args)
	}

	return db.DB.QueryRowContext(ctx, query, args)
}

func (db *Db) Stmt() *sql.Stmt {
	tx = db.tx()
	if tx != nil {
		return tx.Stmt()
	}

	return db.DB.Stmt()
}

func (db *Db) StmtContext(ctx context.Context) *sql.Stmt {
	tx := db.tx()
	if tx != nil {
		return tx.StmtContext(ctx)
	}

	return db.DB.StmtContext(ctx)
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
