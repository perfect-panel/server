package tool

import (
	"testing"
)

func TestFixedUniqueString(t *testing.T) {
	a := "example"
	b := "example1"
	c := "example"

	strA1, err := FixedUniqueString(a, 8, "")
	strB1, err := FixedUniqueString(b, 8, "")
	strC1, err := FixedUniqueString(c, 8, "")
	if err != nil {
		t.Logf("Error: %v", err.Error())
		return
	}
	if strA1 != strC1 {
		t.Errorf("Expected strA1 and strC1 to be equal, got %s and %s", strA1, strC1)
	}
	if strA1 == strB1 {
		t.Errorf("Expected strA1 and strB1 to be different, got %s and %s", strA1, strB1)
	}
	t.Logf("strA1 and strB1 are not equal, strA1: %s, strB1: %s", strA1, strB1)
	t.Logf("strA1 and strC1 are equal,strA1: %s, strC1: %s", strA1, strC1)
}
