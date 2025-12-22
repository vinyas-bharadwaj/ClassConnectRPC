package interceptors

import (
	"context"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type rateLimiter struct {
	mu        sync.Mutex
	visitors  map[string]int
	limit     int
	resetTime time.Duration
}

func NewRateLimiter(limit int, resetTime time.Duration) *rateLimiter {
	rl := &rateLimiter{
		limit:     limit,
		visitors:  make(map[string]int),
		resetTime: resetTime,
	}
	// Starts the reset visitor routine
	go rl.resetVisitorCount()
	return rl
}

func (rl *rateLimiter) resetVisitorCount() {
	for {
		time.Sleep(rl.resetTime)
		// Lock the mutex before clearing visitors
		rl.mu.Lock()
		rl.visitors = map[string]int{}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) RateLimitingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Unable to get client IP")
	}

	visitorIP := p.Addr.String()
	rl.visitors[visitorIP]++
	log.Printf("Visiter count from IP: %s: %d\n", visitorIP, rl.visitors[visitorIP])

	if rl.visitors[visitorIP] > rl.limit {
		return nil, status.Error(codes.ResourceExhausted, "Too many requests")
	}

	return handler(ctx, req)

}
