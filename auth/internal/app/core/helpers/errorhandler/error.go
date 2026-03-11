package errorhandler

import (
	"auth/pkg/graylog/logger"
	"fmt"
	"runtime"
)

func FailOnError(err error, msg string) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		message := fmt.Sprintf("%s: %s", msg, err.Error())
		fmt.Println(message)

		logger.Log.Errorf("Ошибка произошла в файле: %s, строке: %d, сообщение: %s\n", file, line, message)
	}
}

func Fatal(err error, msg string) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		message := fmt.Sprintf("%s: %s", msg, err.Error())
		fmt.Println(message)

		logger.Log.Fatalf("Ошибка произошла в файле: %s, строке: %d, сообщение: %s\n", file, line, message)
	}
}

func LogInfo(message string) {
	logger.Log.Println(message)
	fmt.Println(message)
}
