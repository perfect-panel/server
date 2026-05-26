package traffic

import (
	"testing"
	"time"
)

func TestFirstDayResetAlreadyProcessed(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 30, 0, 0, time.Local)

	tests := []struct {
		name  string
		cache resetTrafficCache
		want  bool
	}{
		{
			name:  "missing cache should not skip",
			cache: resetTrafficCache{},
			want:  false,
		},
		{
			name: "same month should skip",
			cache: resetTrafficCache{
				LastResetTime: time.Date(2026, 6, 1, 0, 31, 0, 0, time.Local),
			},
			want: true,
		},
		{
			name: "previous month should not skip",
			cache: resetTrafficCache{
				LastResetTime: time.Date(2026, 5, 31, 23, 50, 0, 0, time.Local),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstDayResetAlreadyProcessed(now, tt.cache); got != tt.want {
				t.Fatalf("firstDayResetAlreadyProcessed() = %v, want %v", got, tt.want)
			}
		})
	}
}
