package utils

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors   = map[string]*visitor{}
	visitorsMu sync.Mutex
	rlRate     rate.Limit
	rlBurst    int
)

type accountVisitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
	window   time.Duration
}

var (
	accountVisitors   = map[string]*accountVisitor{}
	accountVisitorsMu sync.Mutex
)

func InitRateLimiter() {
	rps := 10.0
	burst := 20
	rlRate = rate.Limit(rps)
	rlBurst = burst
	go func() {
		for {
			time.Sleep(time.Minute)
			visitorsMu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					delete(visitors, ip)
				}
			}
			visitorsMu.Unlock()

			accountVisitorsMu.Lock()
			for key, v := range accountVisitors {
				if time.Since(v.lastSeen) > v.window {
					delete(accountVisitors, key)
				}
			}
			accountVisitorsMu.Unlock()
		}
	}()
}

func AccountActionAllow(action string, accountID uint, window time.Duration) bool {
	return AccountActionAllowBurst(action, accountID, window, 1)
}

func AccountActionAllowBurst(action string, accountID uint, every time.Duration, burst int) bool {
	key := action + ":" + strconv.FormatUint(uint64(accountID), 10)
	return ActionAllowBurst(key, every, burst)
}

func ActionAllowBurst(key string, every time.Duration, burst int) bool {
	if burst < 1 {
		burst = 1
	}
	accountVisitorsMu.Lock()
	defer accountVisitorsMu.Unlock()
	v, ok := accountVisitors[key]
	if !ok {
		lim := rate.NewLimiter(rate.Every(every), burst)
		window := every * time.Duration(burst)
		if window < 3*time.Minute {
			window = 3 * time.Minute
		}
		accountVisitors[key] = &accountVisitor{limiter: lim, lastSeen: time.Now(), window: window}
		return lim.Allow()
	}
	v.lastSeen = time.Now()
	return v.limiter.Allow()
}

func RateLimitAllow(r *http.Request) bool {
	ip := ClientIP(r)
	visitorsMu.Lock()
	defer visitorsMu.Unlock()
	v, ok := visitors[ip]
	if !ok {
		lim := rate.NewLimiter(rlRate, rlBurst)
		visitors[ip] = &visitor{limiter: lim, lastSeen: time.Now()}
		return lim.Allow()
	}
	v.lastSeen = time.Now()
	return v.limiter.Allow()
}
