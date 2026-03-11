package main

import (
	"flag"
	"fmt"
	config "winx-notification/configs"
	"winx-notification/internal/app/core/commands"
	"winx-notification/internal/app/core/helpers/errorhandler"
	"winx-notification/pkg/graylog/logger"
)

func main() {
	signature := flag.String("app", "a app", "a string")
	flag.Parse()

	command, err := commands.MakeCommand(*signature)
	if err != nil {
		errorhandler.Fatal(err, "failed to make command")
	}

	if err := command.Handle(); err != nil {
		errorhandler.Fatal(err, "failed to execute command")
	}

	errorhandler.LogInfo(fmt.Sprintf("command: '%s' executed successfully", *signature))
}

func init() {
	config.InitConfig()
	logger.SetupLogger()
}
