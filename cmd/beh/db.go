package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/gwaycc/log-server/module/db"
	"github.com/gwaylib/errors"
	"github.com/gwaylib/log/behavior"
)

const (
	putBehaviorSql = `
    INSERT INTO
    %s
    (
       event_time, event_key, 
	   req_header, req_params,
       resp_status, resp_params, 
	   use_nsec, uuid
    )VALUES(
            ?, ?, 
			?, ?,
			?, ?,
			?, ?
    )`
	createBehaviorTb = `
CREATE TABLE IF NOT EXISTS %s
(
    event_time DATE NOT NULL,
	-- key format suggest: path_user_ip
    event_key VARCHAR(128) NOT NULL,

	req_header VARCHAR(128),
    req_params BLOB,
    resp_status VARCHAR(10),
    resp_params BLOB,
    use_nsec BIGINT, 

    uuid VARCHAR(64),
    PRIMARY KEY(event_time, uuid),
	KEY(event_key, event_time, resp_status)
) ENGINE=InnoDB;`
)

var (
	behaviorTbName = "beh_"
	behaviorTbTime = "" // 201510
	behaviorTbLock = sync.Mutex{}
)

func getBehaviorTbName(currentTime time.Time) (string, error) {
	behaviorTbLock.Lock()
	defer behaviorTbLock.Unlock()
	timefmt := currentTime.Format("200601")
	if behaviorTbTime == timefmt {
		return behaviorTbName + behaviorTbTime, nil
	}
	newTbName := behaviorTbName + timefmt
	mdb := db.GetCache("master")
	if _, err := mdb.Exec(fmt.Sprintf(createBehaviorTb, newTbName)); err != nil {
		return "", errors.As(err)
	}
	behaviorTbTime = timefmt
	return newTbName, nil
}

func insertBehavior(sha string, l *behavior.Event) error {
	tbName, err := getBehaviorTbName(l.EventTime)
	if err != nil {
		return errors.As(err, l)
	}
	indexLen := len(l.IndexKey)
	if indexLen > 128 {
		l.IndexKey = l.IndexKey[:128]
	}
	mdb := db.GetCache("master")
	_, err = mdb.Exec(fmt.Sprintf(putBehaviorSql, tbName),
		l.EventTime,
		l.IndexKey,
		l.ReqHeader,
		l.ReqParams,
		l.RespStatus,
		l.RespParams,
		l.UseTime,
		sha,
	)
	if err != nil {
		behaviorTbTime = "" // 重新检查表数据是否创建了
		return errors.As(err, l)
	}
	return nil
}
