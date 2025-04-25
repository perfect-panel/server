package tool

import (
	"regexp"
	"strconv"
)

func ExtractVersionNumber(versionStr string) int {
	// 正则表达式匹配括号中的数字 eg. 1.0.0(10000)
	re := regexp.MustCompile(`\((\d+)\)`)
	matches := re.FindStringSubmatch(versionStr)

	if len(matches) > 1 {
		// 转换成数字
		num, err := strconv.Atoi(matches[1])
		if err == nil {
			return num
		}
	}
	return 0
}
