package graylog

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func (m *Message) UnmarshalJSON(data []byte) error {
	i := make(map[string]interface{}, 16)
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	for k, v := range i {
		if k[0] == '_' {
			if m.Extra == nil {
				m.Extra = make(map[string]interface{}, 1)
			}
			m.Extra[k] = v
			continue
		}
		ok := true
		switch k {
		case "version":
			m.Version, ok = v.(string)
		case "host":
			m.Host, ok = v.(string)
		case "short_message":
			m.Short, ok = v.(string)
		case "full_message":
			m.Full, ok = v.(string)
		case "timestamp":
			m.TimeUnix, ok = v.(float64)
		case "level":
			var level float64
			level, ok = v.(float64)
			m.Level = int32(level)
		case "facility":
			m.Facility, ok = v.(string)
		}
		if !ok {
			return fmt.Errorf("invalid type for field %s", k)
		}
	}
	return nil
}
