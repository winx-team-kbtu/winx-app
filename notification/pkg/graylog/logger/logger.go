package logger

import (
	"fmt"
	"winx-notification/configs"
	"winx-notification/pkg/graylog"
	"winx-notification/pkg/graylog/formatter"

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

	if configs.Config.Logger.Host != "" && configs.Config.Logger.Port != 0 {
		tcpWriter, err := graylog.NewTCPWriter(
			fmt.Sprintf("%s:%d", configs.Config.Logger.Host, configs.Config.Logger.Port),
			fmt.Sprintf("%s_%s", configs.Config.App.Environment, configs.Config.Logger.Source),
		)

		if err == nil {
			tcpWriter.Facility = configs.Config.App.Environment
			Log.SetOutput(tcpWriter)
			Log.SetFormatter(new(formatter.GelfFormatter))
			Log.Info("Graylog logging enabled")
		} else {
			Log.Error(fmt.Errorf("failed to setup Graylog: %w", err))
		}
	}

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