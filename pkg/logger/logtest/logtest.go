package logtest

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/perfect-panel/server/pkg/logger"
)

type Buffer struct {
	buf *bytes.Buffer
	t   *testing.T
}

func Discard(t *testing.T) {
	prev := logger.Reset()
	logger.SetWriter(logger.NewWriter(io.Discard))

	t.Cleanup(func() {
		logger.SetWriter(prev)
	})
}

func NewCollector(t *testing.T) *Buffer {
	var buf bytes.Buffer
	writer := logger.NewWriter(&buf)
	prev := logger.Reset()
	logger.SetWriter(writer)

	t.Cleanup(func() {
		logger.SetWriter(prev)
	})

	return &Buffer{
		buf: &buf,
		t:   t,
	}
}

func (b *Buffer) Bytes() []byte {
	return b.buf.Bytes()
}

func (b *Buffer) Content() string {
	var m map[string]interface{}
	if err := json.Unmarshal(b.buf.Bytes(), &m); err != nil {
		return ""
	}

	content, ok := m["content"]
	if !ok {
		return ""
	}

	switch val := content.(type) {
	case string:
		return val
	default:
		// err is impossible to be not nil, unmarshaled from b.buf.Bytes()
		bs, _ := json.Marshal(content)
		return string(bs)
	}
}

func (b *Buffer) Reset() {
	b.buf.Reset()
}

func (b *Buffer) String() string {
	return b.buf.String()
}

func PanicOnFatal(t *testing.T) {
	ok := logger.ExitOnFatal.CompareAndSwap(true, false)
	if !ok {
		return
	}

	t.Cleanup(func() {
		logger.ExitOnFatal.CompareAndSwap(false, true)
	})
}
