package graylog

import (
	"bytes"
	"encoding/json"
	"time"
)

type Message struct {
	Version  string                 `json:"version"`
	Host     string                 `json:"host"`
	Short    string                 `json:"short_message"`
	Full     string                 `json:"full_message,omitempty"`
	TimeUnix float64                `json:"timestamp"`
	Level    int32                  `json:"level,omitempty"`
	Facility string                 `json:"facility,omitempty"`
	Extra    map[string]interface{} `json:"-"`
	RawExtra json.RawMessage        `json:"-"`
}

const (
	LOG_EMERG   = 0
	LOG_ALERT   = 1
	LOG_CRIT    = 2
	LOG_ERR     = 3
	LOG_WARNING = 4
	LOG_NOTICE  = 5
	LOG_INFO    = 6
	LOG_DEBUG   = 7
)

func (m *Message) MarshalJSONBuf(buf *bytes.Buffer) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if _, err = buf.Write(b[:len(b)-1]); err != nil {
		return err
	}
	if len(m.Extra) > 0 {
		eb, err := json.Marshal(m.Extra)
		if err != nil {
			return err
		}
		if err = buf.WriteByte(','); err != nil {
			return err
		}
		if _, err = buf.Write(eb[1 : len(eb)-1]); err != nil {
			return err
		}
	}
	if len(m.RawExtra) > 0 {
		if err := buf.WriteByte(','); err != nil {
			return err
		}
		if _, err = buf.Write(m.RawExtra[1 : len(m.RawExtra)-1]); err != nil {
			return err
		}
	}
	return buf.WriteByte('}')
}

func (m *Message) toBytes(buf *bytes.Buffer) ([]byte, error) {
	if err := m.MarshalJSONBuf(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func constructMessage(p []byte, hostname string, facility string, file string, line int) *Message {
  p = bytes.TrimSpace(p)
  short := p
  full := []byte("")
  if i := bytes.IndexRune(p, '\n'); i > 0 {
    short = p[:i]
    full = p
  }
  extra := map[string]interface{}{}
  json.Unmarshal(short, &extra)

  _ = file
  _ = line

  return &Message{
    Version:  "1.1",
    Host:     hostname,
    Short:    string(short),
    Full:     string(full),
    TimeUnix: float64(time.Now().UnixNano()) / float64(time.Second),
    Level:    6,
    Facility: facility,
    Extra:    extra,
  }
}