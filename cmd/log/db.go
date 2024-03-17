package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/gwaylib/errors"
	"github.com/gwaypg/log-server/module/db"
)

const (
	putLogSql = `
INSERT INTO
%s
(
    md5, platform, version, ip, date, 
    level, logger, msg
)VALUES(
        ?, ?, ?, ?, ?,
        ?, ?, ?
)
`
	createLogTb = `
CREATE TABLE IF NOT EXISTS %s
(
    md5 char(64),

	-- platform name
	platform char(32) NOT NULL,

	-- platform version
	version char(32) NOT NULL,

	-- platform server at
	ip char(64),

	-- log date time
	date DATE NOT NULL,

	-- log level
	level INT NOT NULL,

	-- logger name
	logger char(64) NOT NULL, 

	-- log message
	msg BLOB NOT NULL, 

	PRIMARY KEY (md5),
	KEY (platform,level,date,logger)
);
    `
)

var (
	logTbName = "log_"
	logTbTime = "" // 201510
	logTbLock = sync.Mutex{}
)

func getLogTbName(currentTime time.Time) (string, error) {
	logTbLock.Lock()
	defer logTbLock.Unlock()
	timefmt := currentTime.Format("200601")
	// if the times are same, it has already check the table name.
	if logTbTime == timefmt {
		return logTbName + logTbTime, nil
	}

	tbName := logTbName + timefmt
	mdb := db.GetCache("master")
	// create a new table to storage the log
	if _, err := mdb.Exec(fmt.Sprintf(createLogTb, tbName)); err != nil {
		return "", errors.As(err)
	}
	// set the current table name after create the table successfull.
	logTbTime = timefmt
	return tbName, nil
}

// Add a log to db
func InsertLog(l *DbTable) error {
	tbName, err := getLogTbName(l.date)
	if err != nil {
		return errors.As(err, *l)
	}
	mdb := db.GetCache("master")
	_, err = mdb.Exec(fmt.Sprintf(putLogSql, tbName),
		l.md5,
		l.platform,
		l.version,
		l.ip,
		l.date,
		l.level,
		l.logger,
		l.msg,
	)
	if err != nil {
		logTbTime = "" // when error happend. reinit the table check.
		return errors.As(err, l)
	}
	return nil
}
