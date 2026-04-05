package formatter

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/sirupsen/logrus"
)

type GelfFormatter struct{}

type fields map[string]interface{}

func (f *GelfFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(fields, len(entry.Data)+6)
	blacklist := []string{"_id", "id", "timestamp", "version", "level"}

	for k, v := range entry.Data {
		if contains(k, blacklist) {
			continue
		}
		switch v := v.(type) {
		case error:
			data["_"+k] = v.Error()
		default:
			data["_"+k] = v
		}
	}

	data["version"] = "1.1"
	data["file"] = entry.Caller.File
	data["line"] = entry.Caller.Line
	data["short_message"] = entry.Message
	data["timestamp"] = round((float64(entry.Time.UnixNano())/float64(1000000))/float64(1000), 4)
	data["level"] = toSyslogLevel(entry.Level)
	data["level_name"] = entry.Level.String()
	data["os_hostname"], _ = os.Hostname()

	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}

	return append(serialized, '\n'), nil
}

func contains(needle string, haystack []string) bool {
	for _, a := range haystack {
		if needle == a {
			return true
		}
	}
	return false
}

func round(val float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return math.Floor((val*shift)+.5) / shift
}

func toSyslogLevel(level logrus.Level) int32 {
	switch level {
	case logrus.PanicLevel:
		return 0
	case logrus.FatalLevel:
		return 1
	case logrus.ErrorLevel:
		return 3
	case logrus.WarnLevel:
		return 4
	case logrus.InfoLevel:
		return 6
	case logrus.DebugLevel:
		return 7
	default:
		return 6
	}
}