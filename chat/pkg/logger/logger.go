package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

type Logger struct {
	level  LogLevel
	logger *log.Logger
}

var DefaultLogger *Logger

func InitLogger(level LogLevel, output io.Writer) {
	if output == nil {
		output = os.Stdout
	}

	DefaultLogger = &Logger{
		level:  level,
		logger: log.New(output, "", 0),
	}
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if l.shouldLog(level) {
		msg := fmt.Sprintf(format, args...)
		_, file, line, _ := runtime.Caller(2)
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")

		l.logger.Printf("[%s] %s [%s:%d] %s",
			level, timestamp, file, line, msg)
	}
}

func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LevelDebug: 1,
		LevelInfo:  2,
		LevelWarn:  3,
		LevelError: 4,
	}
	return levels[level] >= levels[l.level]
}

func Debug(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.log(LevelDebug, format, args...)
	}
}

func Info(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.log(LevelInfo, format, args...)
	}
}

func Warn(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.log(LevelWarn, format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.log(LevelError, format, args...)
	}
}

func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		Info("Request: method=%s path=%s status=%d ip=%s latency=%v",
			method, path, statusCode, clientIP, latency)
	}
}

type ContextKey string

const RequestIDKey ContextKey = "request_id"

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, RequestIDKey, id)
}
