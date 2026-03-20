package main

import (
	"flag"
	"fmt"

	config "winx-profile/configs"
	"winx-profile/internal/app/core/commands"
	"winx-profile/internal/app/core/helpers/errorhandler"
	"winx-profile/pkg/graylog/logger"
)

func main() {
	signature := flag.String("app", "noop", "command signature")
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
