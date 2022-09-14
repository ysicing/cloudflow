// AGPL License
// Copyright 2022 ysicing(i@ysicing.me).

package util

import (
	"database/sql"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlMeta struct {
	DBMeta
}

func (mysql MysqlMeta) genDsn() string {
	if mysql.Source.Port == 0 {
		mysql.Source.Port = 3306
	}
	if mysql.Source.User == "" {
		mysql.Source.User = "root"
	}
	return mysql.Source.User + ":" + mysql.Source.Pass + "@tcp(" + mysql.Source.Host + ":" + strconv.Itoa(mysql.Source.Port) + ")/"
}

func (mysql MysqlMeta) CheckAuth() error {
	db, err := sql.Open("mysql", mysql.genDsn())
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Ping()
}
