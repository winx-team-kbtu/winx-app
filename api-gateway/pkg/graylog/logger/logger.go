package logger

import (
	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func init() {
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	Log.SetLevel(logrus.DebugLevel)
}

func SetupLogger() {
	Log.Info("Logger initialized")
}

func LogInfo(message string) {
	Log.Info(message)
}

func LogError(message string, err error) {
	Log.WithField("error", err).Error(message)
}

func LogWarning(message string) {
	Log.Warn(message)
}

func LogDebug(message string) {
	Log.Debug(message)
}

func LogFatal(message string, err error) {
	Log.WithField("error", err).Fatal(message)
}