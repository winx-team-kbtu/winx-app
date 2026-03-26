package middleware

import (
	headercontract "winx-profile/internal/app/core/contracts/microservices/header-contract"
	"winx-profile/internal/app/core/helpers/response"
	"winx-profile/internal/app/models/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AdminOnly(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
			return
		}

		var profile models.Profile
		if err := db.WithContext(ctx.Request.Context()).
			Where("user_id = ?", authUser.ID).
			First(&profile).Error; err != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, response.ErrorResponse("forbidden: admin access required"))
			return
		}

		if profile.Role != models.RoleAdmin {
			ctx.AbortWithStatusJSON(http.StatusForbidden, response.ErrorResponse("forbidden: admin access required"))
			return
		}

		ctx.Next()
	}
}
