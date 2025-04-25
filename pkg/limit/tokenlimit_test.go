package limit

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestTokenLimit_WithCtx(t *testing.T) {
	const (
		total = 100
		rate  = 5
		burst = 10
	)
	store, _ := CreateRedisWithClean(t)
	l := NewTokenLimiter(rate, burst, store, "tokenlimit")

	ctx, cancel := context.WithCancel(context.Background())
	ok := l.AllowCtx(ctx)
	assert.True(t, ok)

	cancel()
	for i := 0; i < total; i++ {
		ok := l.AllowCtx(ctx)
		assert.False(t, ok)
		assert.False(t, l.monitorStarted)
	}
}

func TestTokenLimit_Take(t *testing.T) {
	store, _ := CreateRedisWithClean(t)

	const (
		total = 100
		rate  = 5
		burst = 10
	)
	l := NewTokenLimiter(rate, burst, store, "tokenlimit")
	var allowed int
	for i := 0; i < total; i++ {
		time.Sleep(time.Second / time.Duration(total))
		if l.Allow() {
			allowed++
		}
	}

	assert.True(t, allowed >= burst+rate)
}

func TestTokenLimit_TakeBurst(t *testing.T) {
	store, _ := CreateRedisWithClean(t)

	const (
		total = 100
		rate  = 5
		burst = 10
	)
	l := NewTokenLimiter(rate, burst, store, "tokenlimit")
	var allowed int
	for i := 0; i < total; i++ {
		if l.Allow() {
			allowed++
		}
	}

	assert.True(t, allowed >= burst)
}

// CreateRedisWithClean returns an in process redis.Redis and a clean function.
func CreateRedisWithClean(t *testing.T) (r *redis.Client, clean func()) {
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	}), mr.Close
}
