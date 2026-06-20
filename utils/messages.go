package utils

import (
	"sync/atomic"
	"time"
)

var liveMsgCounter = time.Now().UnixNano()

func NextLiveMessageID() uint {
	return uint(atomic.AddInt64(&liveMsgCounter, 1))
}
