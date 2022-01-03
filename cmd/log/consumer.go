package main

import (
	"context"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/gwaycc/log-server/module/etc"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/log/logger/proto"
	"github.com/gwaylib/redis"
	rmsq "github.com/gwaylib/redis/msq"
)

func Daemon() {
	etcModule := "module/redis"
	rs, err := redis.NewRediStore(
		int(etc.Etc.Int64(etcModule, "pool_size")),
		"tcp",
		etc.Etc.String(etcModule, "uri"),
		etc.Etc.String(etcModule, "passwd"),
	)
	overdue := 5 * time.Minute
	consumer, err := rmsq.NewRedisAutoConsumer(context.TODO(), rmsq.RedisAutoConsumerCfg{
		Redis:         rs,
		StreamName:    etc.Etc.String(etcModule, "stream-name-log"),
		ClaimDuration: overdue,
	})
	if err != nil {
		panic(err)
	}
	// consume
	limit := 10
	for {
		entries, err := consumer.Next(limit, overdue)
		if err != nil {
			if err != redis.ErrNil {
				log.Warn(errors.As(err))
			}
			// the server is still alive, keeping read
			continue
		}
		for _, e := range entries {
			for _, msg := range e.Messages {
				if !handle(&msg) {
					log.Warn(errors.New("handle failed").As(msg))
					continue
				}

				// confirm handle done.
				if err := consumer.ACK(msg.ID); err != nil {
					log.Warn(errors.As(err, msg))
					continue
				}
			}
		}
	}
}

func handle(msg *redis.MessageEntry) bool {
	if len(msg.Fields) != 1 {
		return false
	}
	job := msg.Fields[0]
	p, err := proto.Unmarshal(job.Value)
	if err != nil {
		log.Error(errors.As(err))
		return true
	}

	for _, d := range p.Data {
		msg := ""
		if utf8.Valid(d.Msg) {
			msg = string(d.Msg)
		} else {
			msg = fmt.Sprintf("%#v", d.Msg)
		}
		tb := NewDbTable(
			job.Key,
			p.Context.Platform,
			p.Context.Version,
			p.Context.Ip,
			d.Date.Format(time.RFC3339Nano),
			d.Level.Int(),
			d.Logger,
			msg,
		)
		if err := InsertLog(tb); err != nil {
			log.Error(errors.As(err))
			return true
		}

		// do notify
		switch d.Level {
		case proto.LevelFatal:
			SendMail(tb, 0)
			SendSms(tb, 0)
		case proto.LevelError:
			SendMail(tb, 0)
			SendSms(tb, 0)
		case proto.LevelWarn:
			SendMail(tb, 10)
			SendSms(tb, 50)
		}

	}

	return true
}
