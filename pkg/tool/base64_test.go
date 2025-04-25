package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidImageSize(t *testing.T) {
	testBase64 := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
	maxSize := int64(10)
	result := IsValidImageSize(testBase64, maxSize)
	assert.Equal(t, result, true)
}
