package db

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/gwaylib/conf"
	"github.com/gwaylib/database"
)

var dbFile = conf.RootDir() + "/etc/db.cfg"

func GetCache(section string) *database.DB {
	return database.GetCache(dbFile, section)
}

func HasCache(section string) (*database.DB, error) {
	return database.HasCache(dbFile, section)
}

func CloseCache() {
	database.CloseCache()
}
