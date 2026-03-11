package logger

import (
	"auth/configs"
	"auth/pkg/graylog"
	"auth/pkg/graylog/formatter"
	"fmt"

	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func SetupLogger() {
	var tcpWriter, err = graylog.NewTCPWriter(
		fmt.Sprintf("%s:%d", configs.Config.Logger.Host, configs.Config.Logger.Port),
		fmt.Sprintf("%s_%s", configs.Config.App.Environment, configs.Config.Logger.Source),
	)

	if err == nil {
		tcpWriter.Facility = configs.Config.App.Environment
		Log.SetReportCaller(true)
		Log.SetOutput(tcpWriter)

		Log.Level = logrus.DebugLevel
		Log.SetFormatter(new(formatter.GelfFormatter))
	} else {
		Log.Error(err)
	}
}
