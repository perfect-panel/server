package calculateMonths

import (
	"testing"
	"time"
)

func TestCalculateMonths(t *testing.T) {
	startTime, _ := time.Parse(time.DateTime, "2025-01-15 00:00:00")
	EndTime, _ := time.Parse(time.DateTime, "2025-05-15 00:00:00")
	months := CalculateMonths(startTime, EndTime)
	t.Log(months)
}
