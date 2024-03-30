package util

import "time"

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
