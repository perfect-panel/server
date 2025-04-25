package deduction

import (
	"time"

	"github.com/perfect-panel/server/pkg/tool"
)

const (
	UnitTimeNoLimit = "NoLimit"
	UnitTimeYear    = "Year"
	UnitTimeMonth   = "Month"
	UnitTimeDay     = "Day"
	UintTimeHour    = "Hour"
	UintTimeMinute  = "Minute"

	ResetCycleNone    = 0
	ResetCycle1st     = 1
	ResetCycleMonthly = 2
	ResetCycleYear    = 3
)

type Subscribe struct {
	StartTime      time.Time
	ExpireTime     time.Time
	Traffic        int64
	Download       int64
	Upload         int64
	UnitTime       string
	UnitPrice      int64
	ResetCycle     int64
	DeductionRatio int64
}

type Order struct {
	Amount   int64
	Quantity int64
}

func CalculateRemainingAmount(sub Subscribe, order Order) int64 {
	if sub.UnitTime == UnitTimeNoLimit && sub.ResetCycle != 0 {
		return 0
	}
	// 实际单价
	sub.UnitPrice = order.Amount / order.Quantity
	now := time.Now()
	switch sub.UnitTime {
	case UnitTimeNoLimit:
		usedTraffic := sub.Traffic - sub.Download - sub.Upload
		unitPrice := float64(order.Amount) / float64(sub.Traffic)
		return int64(float64(usedTraffic) * unitPrice)

	case UnitTimeYear:
		remainingYears := tool.YearDiff(now, sub.ExpireTime)
		remainingUnitTimeAmount := calculateRemainingUnitTimeAmount(sub)
		return int64(remainingYears)*sub.UnitPrice + remainingUnitTimeAmount

	case UnitTimeMonth:
		remainingMonths := tool.MonthDiff(now, sub.ExpireTime)
		remainingUnitTimeAmount := calculateRemainingUnitTimeAmount(sub)
		return int64(remainingMonths)*sub.UnitPrice + remainingUnitTimeAmount
	case UnitTimeDay:
		remainingDays := tool.DayDiff(now, sub.ExpireTime)
		remainingUnitTimeAmount := calculateRemainingUnitTimeAmount(sub)
		return remainingDays*sub.UnitPrice + remainingUnitTimeAmount
	}

	return 0
}

func calculateRemainingUnitTimeAmount(sub Subscribe) int64 {
	now := time.Now()
	trafficWeight, timeWeight := calculateWeights(sub.DeductionRatio)
	remainingDays, totalDays := getRemainingAndTotalDays(sub, now)
	remainingTraffic := sub.Traffic - sub.Download - sub.Upload
	remainingTimeAmount := calculateProportionalAmount(sub.UnitPrice, remainingDays, totalDays)
	remainingTrafficAmount := calculateProportionalAmount(sub.UnitPrice, remainingTraffic, sub.Traffic)
	if sub.Traffic == 0 {
		return remainingTimeAmount
	}
	if sub.DeductionRatio != 0 {
		return calculateWeightedAmount(sub.UnitPrice, remainingTraffic, sub.Traffic, remainingDays, totalDays, trafficWeight, timeWeight)
	}

	return min(remainingTimeAmount, remainingTrafficAmount)
}

func calculateWeights(deductionRatio int64) (float64, float64) {
	if deductionRatio == 0 {
		return 0, 0
	}
	trafficWeight := float64(deductionRatio) / 100
	timeWeight := 1 - trafficWeight
	return trafficWeight, timeWeight
}

func getRemainingAndTotalDays(sub Subscribe, now time.Time) (int64, int64) {
	switch sub.ResetCycle {
	case ResetCycleNone:

		remaining := sub.ExpireTime.Sub(now).Hours() / 24
		total := sub.ExpireTime.Sub(sub.StartTime).Hours() / 24
		return int64(remaining), int64(total)

	case ResetCycle1st:
		return tool.DaysToNextMonth(now), tool.GetLastDayOfMonth(now)

	case ResetCycleMonthly:
		// -1 to include the current day
		return tool.DaysToMonthDay(now, sub.StartTime.Day()) - 1, tool.DaysToMonthDay(now, sub.StartTime.Day())
	case ResetCycleYear:
		return tool.DaysToYearDay(now, int(sub.StartTime.Month()), sub.StartTime.Day()),
			tool.GetYearDays(now, int(sub.StartTime.Month()), sub.StartTime.Day())
	}
	return 0, 0
}

func calculateWeightedAmount(unitPrice, remainingTraffic, totalTraffic, remainingDays, totalDays int64, trafficWeight, timeWeight float64) int64 {
	remainingTimeRatio := float64(remainingDays) / float64(totalDays)
	remainingTrafficRatio := float64(remainingTraffic) / float64(totalTraffic)
	weightedRemainingRatio := (timeWeight * remainingTimeRatio) + (trafficWeight * remainingTrafficRatio)
	return int64(float64(unitPrice) * weightedRemainingRatio)
}

func calculateProportionalAmount(unitPrice, remaining, total int64) int64 {
	return int64(float64(unitPrice) * (float64(remaining) / float64(total)))
}
