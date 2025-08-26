package tool

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/perfect-panel/server/pkg/logger"
)

func Int64SliceToStringSlice(slice []int64) []string {
	stringSlice := make([]string, len(slice))
	for i, num := range slice {
		stringSlice[i] = strconv.Itoa(int(num))
	}
	return stringSlice
}
func StringSliceToInt64Slice(slice []string) []int64 {
	int64Slice := make([]int64, len(slice))
	for i, str := range slice {
		num, err := strconv.Atoi(str)
		if err != nil {
			fmt.Println("Error converting string to int:", err)
			continue
		}
		int64Slice[i] = int64(num)
	}
	return int64Slice
}

func StringToInt64Slice(s string) []int64 {

	stringSlice := strings.Split(s, ",")
	var intSlice []int64
	if len(s) == 0 {
		return intSlice
	}
	for _, str := range stringSlice {
		num, err := strconv.ParseInt(strings.TrimSpace(str), 10, 64)
		if err != nil {
			logger.Error("[Tools] StringToInt64Slice",
				logger.Field("error", err.Error()),
				logger.Field("str", str),
			)
			continue
		}
		intSlice = append(intSlice, num)
	}
	return intSlice
}

func Int64SliceToString(intSlice []int64) string {
	var strSlice []string
	for _, num := range intSlice {
		strSlice = append(strSlice, strconv.FormatInt(num, 10))
	}
	return strings.Join(strSlice, ",")
}

// string slice to string
func StringSliceToString(stringSlice []string) string {
	stringSlice = RemoveDuplicateElements(stringSlice...)
	return strings.Join(stringSlice, ",")
}

// StringMergeAndRemoveDuplicates Tool function to convert multiple comma separated strings into [] strings and deduplicate them
func StringMergeAndRemoveDuplicates(strs ...string) []string {
	if len(strs) == 1 && strs[0] == "" {
		return []string{}
	}
	merged := make([]string, 0)
	for _, str := range strs {
		merged = append(merged, strings.Split(str, ",")...)
	}
	uniqueMap := make(map[string]bool)
	var uniqueList []string

	for _, item := range merged {
		if !uniqueMap[item] {
			uniqueMap[item] = true
			uniqueList = append(uniqueList, item)
		}
	}

	return uniqueList
}

func StringSliceContains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func RemoveElementBySlice[T comparable](slice []T, element T) []T {
	result := make([]T, 0, len(slice))
	for _, s := range slice {
		if s != element {
			result = append(result, s)
		}
	}
	return result
}

func RemoveDuplicateElements[T comparable](input ...T) []T {
	uniqueMap := make(map[T]struct{})
	var result []T

	for _, item := range input {
		// 仅在 T 是 string 类型时跳过空字符串
		if v, ok := any(item).(string); ok && v == "" {
			continue
		}
		if _, exists := uniqueMap[item]; !exists {
			uniqueMap[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func Contains[T comparable](slice []T, target T) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}

// RemoveStringElement 移除指定元素
func RemoveStringElement(arr []string, element ...string) []string {
	var result []string
	for _, str := range arr {
		if !Contains(element, str) {
			result = append(result, str)
		}
	}
	return result
}
