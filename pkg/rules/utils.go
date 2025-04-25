package rules

import "strings"

func trimArr(arr []string) []string {
	var result []string
	for _, s := range arr {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
