package handler

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter is a per-client token bucket built on the standard library only.
// Each client IP gets `burst` tokens that refill at `rate` tokens per second;
// a request that finds an empty bucket gets a 429. Idle buckets are swept so
// the map can't grow without bound under churn.
type RateLimiter struct {
	rate  float64 // tokens added per second
	burst float64 // maximum tokens a bucket can hold

	// trustProxy makes clientIP honor Cloudflare's CF-Connecting-IP header.
	// Only enable it when the app sits behind the Cloudflare Tunnel: there the
	// connection always comes from cloudflared (loopback), so without this every
	// request shares one bucket and the per-IP limit becomes global.
	trustProxy bool

	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	tokens   float64
	lastSeen time.Time
}

// NewRateLimiter builds a limiter allowing `burst` requests up front and `rate`
// requests per second sustained, and starts a background sweeper to drop idle
// clients. Set trustProxy when running behind the Cloudflare Tunnel so the
// limiter keys on the real client IP (CF-Connecting-IP) instead of cloudflared.
func NewRateLimiter(rate, burst float64, trustProxy bool) *RateLimiter {
	rl := &RateLimiter{
		rate:       rate,
		burst:      burst,
		trustProxy: trustProxy,
		buckets:    make(map[string]*bucket),
	}
	go rl.sweep()
	return rl
}

// Limit is the middleware: it allows or rejects each request by client IP.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow(clientIP(r, rl.trustProxy)) {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// allow refills the client's bucket based on elapsed time and consumes a token
// if one is available.
func (rl *RateLimiter) allow(ip string) bool {
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[ip]
	if !ok {
		// First request: full bucket, minus the token this request consumes.
		rl.buckets[ip] = &bucket{tokens: rl.burst - 1, lastSeen: now}
		return true
	}

	b.tokens += now.Sub(b.lastSeen).Seconds() * rl.rate
	if b.tokens > rl.burst {
		b.tokens = rl.burst
	}
	b.lastSeen = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// sweep periodically drops buckets that have been idle long enough to have
// refilled completely — they carry no state worth keeping.
func (rl *RateLimiter) sweep() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for now := range ticker.C {
		rl.mu.Lock()
		for ip, b := range rl.buckets {
			if now.Sub(b.lastSeen) > time.Minute {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// clientIP extracts the request's source IP, dropping the port.
//
// When trustProxy is set, it honors Cloudflare's CF-Connecting-IP, which the
// edge overwrites on every request — a single, non-appendable value, unlike the
// multi-hop X-Forwarded-For. This is safe only because the app is reachable
// solely through the tunnel (the port is bound to loopback), so nothing but
// Cloudflare can set that header. With trustProxy off (local dev, direct
// connections) the header is ignored and the connection's RemoteAddr wins,
// since a direct client could spoof it.
func clientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
