package middleware

import (
	"winx-notification/configs"
	"winx-notification/internal/app/core/helpers/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ApiKey() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("x-api-key")
		if token != configs.Config.App.Key {
			ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.UnauthorizedSystem))
			ctx.Abort()

			return
		}

		ctx.Next()
	}
}
