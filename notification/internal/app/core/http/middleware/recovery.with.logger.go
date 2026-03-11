package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"winx-notification/internal/app/core/helpers/response"
	"winx-notification/pkg/graylog/logger"
)

func RecoveryWithLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// stacktrace
				stack := debug.Stack()

				msgf := fmt.Sprintf(
					"panic recovered: stack: %s\n method: %s\n path: %s\n",
					stack,
					c.Request.Method,
					c.Request.URL.Path,
				)

				logger.Log.Error(msgf)
				fmt.Println(msgf)

				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					response.ErrorResponse(response.ServerError),
				)
			}
		}()

		c.Next()
	}
}
