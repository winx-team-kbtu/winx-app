package logger

import (
	"fmt"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func SetupLogger() {
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	Log.SetLevel(logrus.DebugLevel)
	Log.SetReportCaller(true)
	Log.Info("Logger initialized")
}

func LogInfo(message string) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		Log.WithFields(logrus.Fields{
			"file": file,
			"line": line,
		}).Info(message)
	} else {
		Log.Info(message)
	}
	fmt.Printf("[INFO] %s - %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}

func LogError(message string, err error) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		Log.WithFields(logrus.Fields{
			"file":  file,
			"line":  line,
			"error": err,
		}).Error(message)
	} else {
		Log.WithField("error", err).Error(message)
	}
	fmt.Printf("[ERROR] %s - %s: %v\n", time.Now().Format("2006-01-02 15:04:05"), message, err)
}

func LogWarning(message string) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		Log.WithFields(logrus.Fields{
			"file": file,
			"line": line,
		}).Warn(message)
	} else {
		Log.Warn(message)
	}
	fmt.Printf("[WARN] %s - %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}

func LogDebug(message string) {
	if Log.Level >= logrus.DebugLevel {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			Log.WithFields(logrus.Fields{
				"file": file,
				"line": line,
			}).Debug(message)
		} else {
			Log.Debug(message)
		}
		fmt.Printf("[DEBUG] %s - %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
	}
}