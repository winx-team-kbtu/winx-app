package middleware

import (
	headercontract "auth/internal/app/core/contracts/microservices/header-contract"
	"auth/internal/app/core/helpers/errorhandler"
	"auth/internal/app/core/helpers/response"
	"auth/internal/app/models/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AdminOnly(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
			errorhandler.FailOnError(err, "admin middleware: failed to get auth user")
			return
		}

		var user models.User
		if err := db.WithContext(ctx.Request.Context()).
			Preload("Role").
			Where("id = ?", authUser.ID).
			First(&user).Error; err != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, response.ErrorResponse("forbidden: admin access required"))
			errorhandler.FailOnError(err, "admin middleware: failed to get user with role")
			return
		}

		if user.Role == nil || user.Role.Slug != models.RoleSlugAdmin {
			ctx.AbortWithStatusJSON(http.StatusForbidden, response.ErrorResponse("forbidden: admin access required"))
			return
		}

		ctx.Next()
	}
}
