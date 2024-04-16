package util

import (
	"sync"
	"time"
)

// CheckTimestamp ensures the timestamp is in ms
func CheckTimestamp(timestamp int64) int64 {
	if timestamp > 1700000000*1000 { // ms
		return timestamp
	} else {
		return timestamp * 1000
	}
}

// GetTimestamp returns the current timestamp in ms
func GetTimestamp() int64 {
	return time.Now().UnixMilli()
}

// Debounce returns a function that will invoke the provided function after the specified duration
func Debounce(f func(), delay time.Duration) func() {
	var mutex sync.Mutex
	var timer *time.Timer

	return func() {
		mutex.Lock()
		defer mutex.Unlock()

		// stop the previous timer if it's still running
		if timer != nil {
			timer.Stop()
		}

		timer = time.AfterFunc(delay, f)
	}
}
