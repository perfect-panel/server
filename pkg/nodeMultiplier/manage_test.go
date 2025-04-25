package nodeMultiplier

import (
	"testing"
	"time"
)

func TestNewNodeMultiplierManager(t *testing.T) {
	periods := []TimePeriod{
		{
			StartTime:  "23:00",
			EndTime:    "1:59",
			Multiplier: 1.2,
		},
	}
	m := NewNodeMultiplierManager(periods)
	if len(m.Periods) != 1 {
		t.Errorf("expected 1, got %d", len(m.Periods))
	}

	t.Log("00:10 multiplier:", m.GetMultiplier(time.Date(0, 1, 1, 0, 10, 0, 0, time.UTC)))
}
