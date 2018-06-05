package model

import (
	"context"
	"database/sql"
	"strings"
)

// DropRepo will drop the database table about the repo
func (repo *Repo) DropRepo() error {
	repo.Clean()
	tableName := repo.QuotedTableName()
	_, err := repo.exec("DROP TABLE " + tableName)

	return err
}

// CreateRepo will create database table about the repo
func (repo *Repo) CreateRepo() error {
	sqlang, indexes := repo.forCreateTable()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return repo.Tx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(sqlang); err != nil {
			return err
		}
		for _, index := range indexes {
			if _, err := tx.Exec(index); err != nil {
				return err
			}
		}
		return nil
	}, ctx, nil)
}

// forCreateTable generate the create database table sql lang and create database index sql lang
func (repo *Repo) forCreateTable() (sqlang string, indexes []string) {
	sqlang = "CREATE TABLE " + repo.QuotedTableName()
	indexes = []string{}
	cols := []string{}
	repo.model.(Mapable).Mapper().each(func(fd *fieldDescriptor) bool {
		col := []string{fd.colname, fd.coltype}
		if fd.ispk {
			col = append(col, "PRIMARY KEY")
		}
		if !fd.nullable {
			col = append(col, "NOT NULL")
		}
		if fd.isuk {
			tn := repo.model.(Model).TableName()
			index := "CREATE UNIQUE INDEX ui_" + tn + "_" + fd.colname +
				" ON " + tn + " (" + fd.colname + ")"
			indexes = append(indexes, index)
		}
		cols = append(cols, strings.Join(col, " "))
		return true
	})
	sqlang += "(\n\t" + strings.Join(cols, ",\n\t") + "\n)"
	return
}
