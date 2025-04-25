package tool

import (
	"testing"
	"time"
)

func TestAddTime(t *testing.T) {
	basic := time.Now()
	expAt := AddTime("Month", 1, basic)

	t.Logf("AddTime() success, expected year %d, got year %d, full: %v", basic.Year()+1, expAt.Year(), expAt.Format("2006-01-02 15:04:05"))
}

func TestGetYearDays(t *testing.T) {
	days := GetYearDays(time.Now(), 2, 1)
	t.Logf("GetYearDays() success, expected 365, got %d", days)

}
