package db

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/gwaylib/conf"
	"github.com/gwaylib/qsql"
)

var dbFile = conf.RootDir() + "/etc/db.cfg"

func GetCache(section string) *qsql.DB {
	return qsql.GetCache(dbFile, section)
}

func HasCache(section string) (*qsql.DB, error) {
	return qsql.HasCache(dbFile, section)
}

func CloseCache() {
	qsql.CloseCache()
}
