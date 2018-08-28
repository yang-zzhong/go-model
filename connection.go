package model

import (
	"context"
	"database/sql"
	. "github.com/yang-zzhong/go-querybuilder"
)

var (
	Conn     *Connection
	modifier Modifier
	inited   bool
)

// tx callback type
type txhandler func(t *sql.Tx) error

type Connection struct {
	*sql.DB
}

func Config(db *sql.DB, m Modifier) {
	Conn = NewConn(db)
	modifier = m
	inited = true
}

func NewConn(db *sql.DB) *Connection {
	conn := new(Connection)
	conn.DB = db

	return conn
}

func (conn *Connection) Tx(handle txhandler, ctx context.Context, opts *sql.TxOptions) error {
	if ctx == nil {
		nctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctx = nctx
	}
	tx, err := conn.DB.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	defer func() {
		if e := recover(); e != nil {
			tx.Rollback()
			panic(e)
		}
	}()
	if err := handle(tx); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
