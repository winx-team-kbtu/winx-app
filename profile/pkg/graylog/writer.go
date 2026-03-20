package graylog

import (
	"bytes"
	"net"
	"sync"
)

type Writer interface {
	Close() error
	Write([]byte) (int, error)
	WriteMessage(*Message) error
}

type GelfWriter struct {
	addr     string
	conn     net.Conn
	hostname string
	Facility string // defaults to current process name
	proto    string
}

// Close connection and interrupt blocked Read or Write operations
func (w *GelfWriter) Close() error {
	if w.conn == nil {
		return nil
	}
	return w.conn.Close()
}

func newBuffer() *bytes.Buffer {
	b := bufPool.Get().(*bytes.Buffer)
	if b != nil {
		b.Reset()
		return b
	}
	return bytes.NewBuffer(nil)
}

// 1k bytes buffer by default
var bufPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}
