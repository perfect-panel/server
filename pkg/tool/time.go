package tool

import (
	"time"

	"github.com/perfect-panel/server/pkg/logger"
)

func AddTime(unit string, quantity int64, baseTime ...time.Time) time.Time {
	basic := time.Now()
	if len(baseTime) != 0 {
		basic = baseTime[0]
	}
	switch unit {
	case "Year":
		return basic.AddDate(int(quantity), 0, 0)
	case "Month":
		return basic.AddDate(0, int(quantity), 0)
	case "Day":
		return basic.AddDate(0, 0, int(quantity))
	case "Hour":
		return basic.Add(time.Hour * time.Duration(quantity))
	case "Minute":
		return basic.Add(time.Minute * time.Duration(quantity))
	case "NoLimit":
		return time.UnixMilli(0)
	default:
		logger.Error("[Tool] Unknown time unit", logger.Field("unit", unit))
		return basic
	}
}

func MonthDiff(startTime, endTime time.Time) int {
	startYear, startMonth, startDay := startTime.Year(), int(startTime.Month()), startTime.Day()
	endYear, endMonth, endDay := endTime.Year(), int(endTime.Month()), endTime.Day()

	// 计算初步月份差
	monthDiff := (endYear-startYear)*12 + (endMonth - startMonth)

	// 检查是否要扣除不足一个完整月的部分
	if endDay <= startDay {
		monthDiff-- // 如果结束时间的日期小于开始时间的日期，则不计为完整自然月
	}

	return monthDiff
}
func DaysToMonthDay(t time.Time, targetDay int) int64 {
	currentDay := t.Day()
	year, month := t.Year(), t.Month()

	var targetDate time.Time

	// 如果当前号数大于目标号数，计算本月目标号数
	if currentDay > targetDay {
		targetDate = getValidDate(year, month, targetDay, t.Location())
	} else { // 如果当前号数小于等于目标号数，计算上个月目标号数
		if month == time.January {
			year--
			month = time.December
		} else {
			month--
		}
		targetDate = getValidDate(year, month, targetDay, t.Location())
	}

	// 计算时间差
	duration := t.Sub(targetDate)
	return int64(duration.Hours() / 24) // 转换为整天数
}

func DaysToNextMonth(t time.Time) int64 {
	// 获取下个月的1号
	year, month := t.Year(), t.Month()
	if month == 12 {
		year++
		month = 1
	} else {
		month++
	}
	nextMonthFirstDay := time.Date(year, month, 1, 0, 0, 0, 0, t.Location())

	// 计算时间差
	duration := nextMonthFirstDay.Sub(t)
	return int64(duration.Hours() / 24) // 转换为整天数
}

func getValidDate(year int, month time.Month, day int, loc *time.Location) time.Time {
	// 构造当月的 1 号
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, loc)

	// 获取当月的天数
	lastDayOfMonth := firstOfMonth.AddDate(0, 1, -1).Day()

	// 如果目标号数超过当月的天数，则使用最后一天
	if day > lastDayOfMonth {
		day = lastDayOfMonth
	}

	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

// GetLastDayOfMonth 获取指定时间所在月份的最后一天
func GetLastDayOfMonth(t time.Time) int64 {
	// 获取当前月份的第一天
	firstDayOfMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	// 获取下个月的第一天，然后减去一天
	lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, 0).Add(-time.Hour * 24)
	return int64(lastDayOfMonth.Day())
}

func GetYearDays(t time.Time, month int, day int) int64 {
	startTime := time.Date(t.Year(), time.Month(month), day, 0, 0, 0, 0, t.Location())
	endTime := time.Date(t.Year(), time.Month(month), day, 0, 0, 0, 0, t.Location())
	if endTime.After(t) {
		startTime = time.Date(t.Year()-1, time.Month(month), day, 0, 0, 0, 0, t.Location())
	} else {
		endTime = time.Date(t.Year()+1, time.Month(month), day, 0, 0, 0, 0, t.Location())
	}
	return int64(endTime.Sub(startTime).Hours() / 24)
}

func DaysToYearDay(t time.Time, month int, day int) int64 {
	targetTime := time.Date(t.Year(), time.Month(month), day, 0, 0, 0, 0, t.Location())
	if targetTime.Before(t) {
		targetTime = time.Date(t.Year()+1, time.Month(month), day, 0, 0, 0, 0, t.Location())
	}

	return int64(t.Sub(targetTime).Hours() / 24)
}

func YearDiff(startTime, endTime time.Time) int {
	// 计算基础年份差
	yearDiff := endTime.Year() - startTime.Year()

	// 检查结束时间是否在开始时间之前（同一年的同一天之前）
	if endTime.Month() < startTime.Month() || (endTime.Month() == startTime.Month() && endTime.Day() < startTime.Day()) {
		yearDiff-- // 不足一年则减去一年
	}
	return yearDiff
}

func DayDiff(startTime, endTime time.Time) int64 {
	// 计算时间差
	duration := endTime.Sub(startTime)
	return int64(duration.Hours() / 24) // 转换为整天数
}

// HourDiff 计算两个时间点之间的小时差
func HourDiff(startTime, endTime time.Time) int64 {
	// 计算时间差
	duration := endTime.Sub(startTime)
	return int64(duration.Hours()) // 返回小时数，可能包含小数部分
}
