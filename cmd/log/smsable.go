package main

import (
	"sync"
	"time"
)

// TODO: every 5 minute cheking the timeout, if it's timeout, need recycling.
type smsAble struct {
	beginTime time.Time
	times     int
}

var (
	smsAbleLk    = sync.Mutex{}
	smsAbleTimes = make(map[string]*smsAble)
)

// count the sms send times, and limit the send.
func isSmsAble(platform string, minTimes int, now time.Time, timeout time.Duration) bool {
	smsAbleLk.Lock()
	defer smsAbleLk.Unlock()

	sAble, ok := smsAbleTimes[platform]
	if !ok {
		sAble = &smsAble{
			beginTime: now,
			times:     1,
		}
		smsAbleTimes[platform] = sAble
	} else {
		// checking the time
		d := now.Sub(sAble.beginTime)
		if d < timeout {
			sAble.times++
		} else {
			// timeout, recount the send times.
			sAble.beginTime = now
			sAble.times = 1
		}
	}

	// The minimum number of times to send condition is not reached
	if sAble.times < minTimes {
		return false
	}

	// the first send.
	if sAble.times == minTimes {
		return true
	}

	return false
}
