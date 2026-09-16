package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"football-app/pkg/apperror"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter returns a per-client-IP token-bucket rate limiter: rps is the
// sustained requests-per-second allowed, burst is the max immediate burst.
// State is in-memory and per-process — correct for a single instance; a
// deployment with multiple replicas behind a load balancer would need a
// shared store (e.g. Redis) instead, since each replica would otherwise
// track its own independent limit.
func RateLimiter(rps float64, burst int) gin.HandlerFunc {
	var (
		mu       sync.Mutex
		visitors = make(map[string]*visitor)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	getLimiter := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		v, ok := visitors[ip]
		if !ok {
			l := rate.NewLimiter(rate.Limit(rps), burst)
			visitors[ip] = &visitor{limiter: l, lastSeen: time.Now()}
			return l
		}
		v.lastSeen = time.Now()
		return v.limiter
	}

	return func(c *gin.Context) {
		if !getLimiter(c.ClientIP()).Allow() {
			c.Error(apperror.TooManyRequests("too many requests, please try again later"))
			c.Abort()
			return
		}
		c.Next()
	}
}
