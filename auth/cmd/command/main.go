package main

import (
	config "auth/configs"
	"auth/internal/app/core/commands"
	"auth/internal/app/core/helpers/errorhandler"
	"auth/pkg/graylog/logger"
	"flag"
	"fmt"
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
