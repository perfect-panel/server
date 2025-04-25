package limit

import (
	"testing"

	"github.com/redis/go-redis/v9"

	"github.com/stretchr/testify/assert"
)

func TestPeriodLimit_Take(t *testing.T) {
	testPeriodLimit(t)
}

func TestPeriodLimit_TakeWithAlign(t *testing.T) {
	testPeriodLimit(t, Align())
}

func TestPeriodLimit_RedisUnavailable(t *testing.T) {
	//t.Skipf("skip this test because it's not stable")
	const (
		seconds = 1
		quota   = 5
	)
	rds := redis.NewClient(&redis.Options{
		Addr: "localhost:12345",
	})

	l := NewPeriodLimit(seconds, quota, rds, "periodlimit:")
	val, err := l.Take("first")
	assert.NotNil(t, err)
	assert.Equal(t, 0, val)
}

func testPeriodLimit(t *testing.T, opts ...PeriodOption) {
	store, _ := CreateRedisWithClean(t)
	const (
		seconds = 1
		total   = 100
		quota   = 5
	)
	l := NewPeriodLimit(seconds, quota, store, "periodlimit", opts...)
	var allowed, hitQuota, overQuota int
	for i := 0; i < total; i++ {
		val, err := l.Take("first")
		if err != nil {
			t.Error(err)
		}
		switch val {
		case Allowed:
			allowed++
		case HitQuota:
			hitQuota++
		case OverQuota:
			overQuota++
		default:
			t.Error("unknown status")
		}
	}
	assert.Equal(t, quota-1, allowed)
	assert.Equal(t, 1, hitQuota)
	assert.Equal(t, total-quota, overQuota)
}

func TestQuotaFull(t *testing.T) {
	rds, _ := CreateRedisWithClean(t)
	l := NewPeriodLimit(1, 1, rds, "periodlimit")
	val, err := l.Take("first")
	assert.Nil(t, err)
	assert.Equal(t, HitQuota, val)
}
