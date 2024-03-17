package main

import (
	"context"
	"time"

	"github.com/gwaypg/log-server/module/etc"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/log/behavior"
	"github.com/gwaylib/log/logger"
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
		StreamName:    etc.Etc.String(etcModule, "stream-name-beh"),
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
	event, err := behavior.Parse(job.Value)
	if err != nil {
		logger.FailLog(err)
		return true
	}
	if err := insertBehavior(job.Key, event); err != nil {
		logger.FailLog(err)
		return true
	}
	return true
}
