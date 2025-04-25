package logtest

import (
	"errors"
	"testing"

	"github.com/perfect-panel/server/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestCollector(t *testing.T) {
	const input = "hello"
	c := NewCollector(t)
	logger.Info(input)
	assert.Equal(t, input, c.Content())
	assert.Contains(t, c.String(), input)
	c.Reset()
	assert.Empty(t, c.Bytes())
}

func TestPanicOnFatal(t *testing.T) {
	const input = "hello"
	Discard(t)
	logger.Info(input)

	PanicOnFatal(t)
	PanicOnFatal(t)
	assert.Panics(t, func() {
		logger.Must(errors.New("foo"))
	})
}

func TestCollectorContent(t *testing.T) {
	const input = "hello"
	c := NewCollector(t)
	c.buf.WriteString(input)
	assert.Empty(t, c.Content())
	c.Reset()
	c.buf.WriteString(`{}`)
	assert.Empty(t, c.Content())
	c.Reset()
	c.buf.WriteString(`{"content":1}`)
	assert.Equal(t, "1", c.Content())
}
