package nodeMultiplier

import "time"

type TimePeriod struct {
	StartTime  string  `json:"start_time"`
	EndTime    string  `json:"end_time"`
	Multiplier float32 `json:"multiplier"`
}

type Manager struct {
	Periods []TimePeriod
}

func NewNodeMultiplierManager(periods []TimePeriod) *Manager {
	return &Manager{
		Periods: periods,
	}
}

func (m *Manager) GetMultiplier(current time.Time) float32 {
	for _, period := range m.Periods {
		if m.isInTimePeriod(current, period.StartTime, period.EndTime) {
			return period.Multiplier
		}
	}
	return 1 // Default multiplier is 1 (no change)
}

func (m *Manager) isInTimePeriod(current time.Time, start, end string) bool {
	startTime, _ := time.Parse("15:04.000", start)
	endTime, _ := time.Parse("15:04.000", end)

	currentTime := time.Date(0, 1, 1, current.Hour(), current.Minute(), 0, 0, time.UTC)
	startTimeFormatted := time.Date(0, 1, 1, startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
	endTimeFormatted := time.Date(0, 1, 1, endTime.Hour(), endTime.Minute(), 0, 0, time.UTC)

	if startTimeFormatted.Before(endTimeFormatted) {
		return currentTime.After(startTimeFormatted) && currentTime.Before(endTimeFormatted)
	}
	// Handle ranges that cross midnight
	return currentTime.After(startTimeFormatted) || currentTime.Before(endTimeFormatted)
}
