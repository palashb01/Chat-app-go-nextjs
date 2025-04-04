package main

import (
    "net/http"
    "sync"

    "golang.org/x/time/rate"
)

// ClientLimiter manages rate limiting per client
type ClientLimiter struct {
    clients map[string]*rate.Limiter
    mu      sync.Mutex
}

// NewClientLimiter initializes a rate limiter map
func NewClientLimiter() *ClientLimiter {
    return &ClientLimiter{
        clients: make(map[string]*rate.Limiter),
    }
}

// GetLimiter returns the rate limiter for a given IP
func (cl *ClientLimiter) GetLimiter(ip string) *rate.Limiter {
    cl.mu.Lock()
    defer cl.mu.Unlock()

    if _, exists := cl.clients[ip]; !exists {
        cl.clients[ip] = rate.NewLimiter(10, 5) // 1 request per second, burst of 5
    }
    return cl.clients[ip]
}

var limiter = NewClientLimiter()

// RateLimitMiddleware applies rate limiting
func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := r.RemoteAddr
        if !limiter.GetLimiter(ip).Allow() {
            http.Error(w, "Too many requests", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
