package main

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gwaylib/errors"
)

const (
	// some database need to escape to save
	// mysql true
	// pg	false
	magicQuotesOn = true
)

type DbTable struct {
	md5      string
	platform string
	version  string
	ip       string
	date     time.Time
	level    int
	logger   string
	msg      string
}

func NewDbTable(key, platform, ver, ip, date string, level int, logger, msg string) *DbTable {
	d, err := time.Parse(time.RFC3339, date)
	if err != nil {
		SendLogFail(errors.As(err, date))
	}
	return &DbTable{
		key,
		platform,
		ver,
		ip,
		d,
		level,
		logger,
		msg,
	}
}

func (dt *DbTable) MD5() string {
	return dt.md5
}
func (dt *DbTable) Platform() string {
	return dt.platform
}
func (dt *DbTable) Version() string {
	return dt.version
}
func (dt *DbTable) Ip() string {
	return dt.ip
}
func (dt *DbTable) Date() string {
	return dt.date.Format(time.RFC3339Nano)
}
func (dt *DbTable) Level() int {
	return dt.level
}
func (dt *DbTable) Logger() string {
	return dt.logger
}
func (dt *DbTable) Msg() string {
	if magicQuotesOn {
		return strings.Replace(dt.msg, "\\", "\\\\", -1)
	}
	return dt.msg
}

func (dt *DbTable) MsgIndent() string {
	var msg map[string]interface{}
	if err := json.Unmarshal([]byte(dt.msg), &msg); err != nil {
		return dt.msg
	}
	data, err := json.MarshalIndent(&msg, "", "	")
	if err != nil {
		return dt.msg
	}

	return string(data)
}
