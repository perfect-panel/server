package logger

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

func getCaller(callDepth int) string {
	var file string
	var line int
	var ok bool
	noMatch := []string{"logger", "@", "model", "default"}

	if callDepth > 0 {
		_, file, line, ok = runtime.Caller(callDepth)
		if !ok {
			return ""
		}
	} else {
		// skip logger and External library
		for i := 0; i < 20; i++ {
			_, file, line, ok = runtime.Caller(i)
			if !ok {
				return ""
			}
			if !contains(noMatch, file) {
				break
			}
		}
	}
	return prettyCaller(file, line)
}

func getTimestamp() string {
	return time.Now().Format(timeFormat)
}

func prettyCaller(file string, line int) string {
	idx := strings.LastIndexByte(file, '/')
	if idx < 0 {
		return fmt.Sprintf("%s:%d", file, line)
	}

	idx = strings.LastIndexByte(file[:idx], '/')
	if idx < 0 {
		return fmt.Sprintf("%s:%d", file, line)
	}

	return fmt.Sprintf("%s:%d", file[idx+1:], line)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.Contains(item, s) {
			return true
		}
	}
	return false
}
