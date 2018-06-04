package model

import (
	"context"
	"database/sql"
)

// tx callback type
type txhandler func(t *sql.Tx) error

// tell repo use tx to exec sql
func (repo *Repo) WithTx(tx *sql.Tx) *Repo {
	repo.tx = tx
	return repo
}

// tell repo use conn to exec sql
func (repo *Repo) WithoutTx() *Repo {
	repo.tx = nil
	return repo
}

// exec trans
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
