package model

import (
	"context"
	"database/sql"
)

type txhandler func(t *sql.Tx) error

func (repo *Repo) WithTx(tx *sql.Tx) *Repo {
	repo.tx = tx
	return repo
}

func (repo *Repo) WithoutTx() *Repo {
	repo.tx = nil
	return repo
}

func (repo *Repo) Tx(handle txhandler, ctx context.Context, opts *sql.TxOptions) error {
	tx, err := repo.conn.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	if err := handle(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
