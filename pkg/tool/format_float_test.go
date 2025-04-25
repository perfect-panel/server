package tool

import "testing"

func TestFormatStringToFloat(t *testing.T) {
	var value = 1.23
	if FormatStringToFloat("1.23") != value {
		t.Errorf("Expected %f, but got %f", value, FormatStringToFloat("1.23"))
	}
}
