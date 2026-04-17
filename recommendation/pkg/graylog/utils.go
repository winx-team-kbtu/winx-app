package graylog

import (
	"runtime"
	"strings"
)

func getCaller(callDepth int, suffixesToIgnore ...string) (file string, line int) {
	callDepth++
outer:
	for {
		var ok bool
		_, file, line, ok = runtime.Caller(callDepth)
		if !ok {
			file = "???"
			line = 0
			break
		}

		for _, s := range suffixesToIgnore {
			if strings.HasSuffix(file, s) {
				callDepth++
				continue outer
			}
		}
		break
	}
	return
}

func getCallerIgnoringLogMulti(callDepth int) (string, int) {
	return getCaller(callDepth+1, "/pkg/log/log.go", "/pkg/io/multi.go")
}
