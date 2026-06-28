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
			StartTime:  "12:00",
			EndTime:    "13:59",
			Multiplier: 0.5,
		},
	}
	m := NewNodeMultiplierManager(periods)
	if len(m.Periods) != 2 {
		t.Fatalf("periods len = %d, want 2", len(m.Periods))
	}

	tests := []struct {
		name string
		now  time.Time
		want float32
	}{
		{
			name: "cross midnight",
			now:  time.Date(0, 1, 1, 0, 10, 0, 0, time.UTC),
			want: 1.2,
		},
		{
			name: "daytime",
			now:  time.Date(0, 1, 1, 12, 30, 0, 0, time.UTC),
			want: 0.5,
		},
		{
			name: "outside periods",
			now:  time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC),
			want: 1,
		},
		{
			name: "html time value",
			now:  time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC),
			want: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.GetMultiplier(tt.now)
			if got != tt.want {
				t.Fatalf("GetMultiplier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeMultiplierInvalidPeriod(t *testing.T) {
	m := NewNodeMultiplierManager([]TimePeriod{
		{
			StartTime:  "invalid",
			EndTime:    "13:59",
			Multiplier: 0.5,
		},
	})

	got := m.GetMultiplier(time.Date(0, 1, 1, 12, 30, 0, 0, time.UTC))
	if got != 1 {
		t.Fatalf("GetMultiplier() = %v, want default multiplier 1", got)
	}
}
