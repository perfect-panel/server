package nodeMultiplier

import (
	"testing"
	"time"
)

func TestNewNodeMultiplierManager(t *testing.T) {
	periods := []TimePeriod{
		{
			StartTime:  "23:00.000",
			EndTime:    "1:59.000",
			Multiplier: 1.2,
		},
		{
			StartTime:  "12:00.000",
			EndTime:    "13:59.000",
			Multiplier: 0.5,
		},
	}
	m := NewNodeMultiplierManager(periods)
	if len(m.Periods) != 1 {
		t.Errorf("expected 1, got %d", len(m.Periods))
	}

	t.Log("00:10 multiplier:", m.GetMultiplier(time.Date(0, 1, 1, 0, 10, 0, 0, time.UTC)))
}
