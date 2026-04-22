package middleware

import (
	"time"

	"winx-profile/pkg/graylog/logger"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Log.WithFields(logrus.Fields{
			"method":    c.Request.Method,
			"path":      c.Request.URL.Path,
			"status":    c.Writer.Status(),
			"latency":   time.Since(start).String(),
			"client_ip": c.ClientIP(),
		}).Info("http request")
	}
}
