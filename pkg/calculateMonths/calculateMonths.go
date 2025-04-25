package calculateMonths

import "time"

// CalculateMonths calculates the number of months between startTime and endTime.
// It rounds up to the next month if there are remaining days.
func CalculateMonths(startTime, endTime time.Time) int8 {
	// Calculate the year and month difference
	years := endTime.Year() - startTime.Year()
	months := int8(years*12) + int8(endTime.Month()) - int8(startTime.Month())

	// Always round up if endTime is not on the same or earlier day of the month
	if endTime.Day() > startTime.Day() || (endTime.Day() < startTime.Day() && endTime.After(startTime)) {
		months++
	}
	return months
}
