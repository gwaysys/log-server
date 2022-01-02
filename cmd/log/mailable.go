package main

import (
	"sync"
	"time"
)

type mailAble struct {
	beginTime time.Time
	times     int
}

var (
	mailAbleLk    = sync.Mutex{}
	mailAbleTimes = make(map[string]*mailAble)
)

func isMailAble(platform string, minTimes int, now time.Time, cacheDuration time.Duration) bool {
	mailAbleLk.Lock()
	defer mailAbleLk.Unlock()

	sAble, ok := mailAbleTimes[platform]
	if !ok {
		sAble = &mailAble{
			beginTime: now,
			times:     1,
		}
		mailAbleTimes[platform] = sAble
	} else {
		d := now.Sub(sAble.beginTime)
		if d < cacheDuration {
			sAble.times++
		} else {
			sAble.beginTime = now
			sAble.times = 1
		}
	}

	if sAble.times < minTimes {
		return false
	}

	if sAble.times == minTimes {
		return true
	}

	return false
}
