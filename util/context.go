package util

import (
	"context"
	"time"
)

func GetTimeoutContext(second int) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(second)*time.Second)
}
