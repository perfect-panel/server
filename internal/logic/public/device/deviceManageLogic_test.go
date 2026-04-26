package device

// 重置频率限制纯函数单测。
// 不依赖 DB / Redis,所以可以在任何环境跑。

import (
	"strings"
	"testing"
	"time"

	"github.com/perfect-panel/server/internal/model/user"
)

func TestResetFrequency_AllowsUnderHourLimit(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	d := &user.SubscribeDevice{
		ResetCountHour: 2, // < 3
		ResetCountDay:  5, // < 10
		LastResetAt:    pTime(now.Add(-10 * time.Minute)),
	}
	if err := applyResetFrequencyCheck(d, now); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestResetFrequency_BlocksAtHourLimit(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 30, 0, 0, time.UTC)
	d := &user.SubscribeDevice{
		ResetCountHour: 3, // == 3 = blocked (决策 9: < 3)
		ResetCountDay:  3,
		LastResetAt:    pTime(now.Add(-1 * time.Minute)),
	}
	err := applyResetFrequencyCheck(d, now)
	if err == nil {
		t.Fatal("expected hour-rate-limit error, got nil")
	}
	if !strings.Contains(err.Error(), "3/h") {
		t.Errorf("error should mention 3/h, got: %v", err)
	}
}

func TestResetFrequency_BlocksAtDayLimit(t *testing.T) {
	now := time.Date(2026, 5, 1, 23, 0, 0, 0, time.UTC)
	d := &user.SubscribeDevice{
		ResetCountHour: 1,
		ResetCountDay:  10, // == 10 = blocked
		LastResetAt:    pTime(now.Add(-30 * time.Minute)),
	}
	err := applyResetFrequencyCheck(d, now)
	if err == nil {
		t.Fatal("expected day-rate-limit error, got nil")
	}
	if !strings.Contains(err.Error(), "10/d") {
		t.Errorf("error should mention 10/d, got: %v", err)
	}
}

func TestResetFrequency_AutoZeroOnNewHour(t *testing.T) {
	last := time.Date(2026, 5, 1, 11, 30, 0, 0, time.UTC)
	now := time.Date(2026, 5, 1, 12, 5, 0, 0, time.UTC) // crossed hour boundary
	d := &user.SubscribeDevice{
		ResetCountHour: 3, // would block, but cross-hour zeroes it
		ResetCountDay:  4,
		LastResetAt:    &last,
	}
	if err := applyResetFrequencyCheck(d, now); err != nil {
		t.Errorf("cross-hour should auto-zero hour counter; got %v", err)
	}
	if d.ResetCountHour != 0 {
		t.Errorf("expected hour counter zeroed, got %d", d.ResetCountHour)
	}
	if d.ResetCountDay != 4 {
		t.Errorf("expected day counter unchanged at 4, got %d", d.ResetCountDay)
	}
}

func TestResetFrequency_AutoZeroOnNewDay(t *testing.T) {
	last := time.Date(2026, 5, 1, 23, 50, 0, 0, time.UTC)
	now := time.Date(2026, 5, 2, 0, 5, 0, 0, time.UTC) // next day
	d := &user.SubscribeDevice{
		ResetCountHour: 2,
		ResetCountDay:  10, // would block, but cross-day zeroes both
		LastResetAt:    &last,
	}
	if err := applyResetFrequencyCheck(d, now); err != nil {
		t.Errorf("cross-day should auto-zero both counters; got %v", err)
	}
	if d.ResetCountHour != 0 || d.ResetCountDay != 0 {
		t.Errorf("expected both counters zeroed, got hour=%d day=%d", d.ResetCountHour, d.ResetCountDay)
	}
}
