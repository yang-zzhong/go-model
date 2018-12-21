package model

import (
	"strings"
)

func (repo *Repo) DropRepoDB(db DB) error {
	repo.Clean()
	tableName := repo.QuotedTableName()
	_, err := db.Exec("DROP TABLE " + tableName)

	return err
}

// DropRepo will drop the database table about the repo
func (repo *Repo) DropRepo() error {
	return repo.DropRepoDB(GetDefaultDB())
}

func (repo *Repo) CreateRepo() error {
	return repo.CreateRepoDB(GetDefaultDB())
}

// CreateRepo will create database table about the repo
func (repo *Repo) CreateRepoDB(db DB) error {
	sqlang, indexes := repo.forCreateTable()
	if _, err := db.Exec(sqlang, repo.Params()...); err != nil {
		return err
	}
	for _, index := range indexes {
		if _, err := db.Exec(index); err != nil {
			return err
		}
	}
	return nil
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
