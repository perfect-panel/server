package logger

import (
	"testing"
)

func TestReadLastNLines(t *testing.T) {
	t.Skipf("skip this test until this test fails")
	lines, err := ReadLastNLines("/Users/tension/code/ppanel/server/logs", 10)
	if err != nil {
		t.Fatalf("Error reading last N lines: %v", err)
	}
	for i, line := range lines {
		t.Logf("Line %d: %s", i, line)
	}
}
