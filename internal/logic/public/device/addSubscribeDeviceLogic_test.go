package device

// 加购单设备 — 按比例计费纯函数单测。
// 不依赖 DB / Redis,所以可以在任何环境跑。

import (
	"testing"
	"time"

	subscribepkg "github.com/perfect-panel/server/internal/model/subscribe"
	"github.com/perfect-panel/server/internal/model/user"
)

func TestProRatedDevicePrice_FullPeriodReturnsFullPrice(t *testing.T) {
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	sub := &user.Subscribe{
		StartTime:  now,
		ExpireTime: now.AddDate(0, 0, 30),
	}
	plan := &subscribepkg.Subscribe{UnitPricePerDevice: 1000}
	amount, ratio := proRatedDevicePrice(plan, sub, now)
	if amount != 1000 {
		t.Errorf("expected amount=1000, got %d", amount)
	}
	if ratio != 10000 {
		t.Errorf("expected ratio=10000 (100%%), got %d", ratio)
	}
}

func TestProRatedDevicePrice_HalfPeriod(t *testing.T) {
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	now := start.AddDate(0, 0, 15)
	sub := &user.Subscribe{
		StartTime:  start,
		ExpireTime: start.AddDate(0, 0, 30),
	}
	plan := &subscribepkg.Subscribe{UnitPricePerDevice: 1000}
	amount, ratio := proRatedDevicePrice(plan, sub, now)
	if amount != 500 {
		t.Errorf("half period: expected 500, got %d", amount)
	}
	if ratio != 5000 {
		t.Errorf("half period: expected ratio 5000, got %d", ratio)
	}
}

func TestProRatedDevicePrice_ExpiredReturnsZero(t *testing.T) {
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	now := start.AddDate(0, 1, 0) // a month later, expired
	sub := &user.Subscribe{
		StartTime:  start,
		ExpireTime: start.AddDate(0, 0, 30),
	}
	plan := &subscribepkg.Subscribe{UnitPricePerDevice: 1000}
	amount, ratio := proRatedDevicePrice(plan, sub, now)
	if amount != 0 || ratio != 0 {
		t.Errorf("expired: expected (0,0), got (%d,%d)", amount, ratio)
	}
}

func TestProRatedDevicePrice_DocExample(t *testing.T) {
	// DoD #2:30 天用了 10 天还剩 20 天,加购 1 设备 = 67% 单价。
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	now := start.AddDate(0, 0, 10)
	sub := &user.Subscribe{
		StartTime:  start,
		ExpireTime: start.AddDate(0, 0, 30),
	}
	plan := &subscribepkg.Subscribe{UnitPricePerDevice: 1000} // ¥10
	amount, ratio := proRatedDevicePrice(plan, sub, now)
	// 20/30 ≈ 0.6667 → ¥10 × 0.6667 = 666.67 → round → 667
	if amount != 667 {
		t.Errorf("doc example: expected 667 (rounded from 666.67), got %d", amount)
	}
	if ratio != 6667 {
		t.Errorf("doc example: expected ratio bp 6667, got %d", ratio)
	}
}

// pTime — helper for *time.Time literals (shared with deviceManageLogic_test.go).
func pTime(t time.Time) *time.Time {
	return &t
}
