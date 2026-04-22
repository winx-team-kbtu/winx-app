package logger

import (
	"fmt"
	"time"
	"winx-profile/configs"
	"winx-profile/pkg/graylog"
	"winx-profile/pkg/graylog/formatter"

	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func SetupLogger() {
	addr := fmt.Sprintf("%s:%d", configs.Config.Logger.Host, configs.Config.Logger.Port)
	hostname := fmt.Sprintf("%s_%s", configs.Config.App.Environment, configs.Config.Logger.Source)

	go func() {
		for i := 0; ; i++ {
			tcpWriter, err := graylog.NewTCPWriter(addr, hostname)
			if err == nil {
				tcpWriter.Facility = configs.Config.App.Environment
				Log.SetReportCaller(true)
				Log.SetOutput(tcpWriter)
				Log.Level = logrus.DebugLevel
				Log.SetFormatter(new(formatter.GelfFormatter))
				fmt.Println("graylog connected")
				return
			}
			fmt.Printf("graylog connect attempt %d failed: %s, retrying in 5s\n", i+1, err)
			time.Sleep(5 * time.Second)
		}
	}()
}
