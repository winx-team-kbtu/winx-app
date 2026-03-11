package middleware

import (
	headercontract "auth/internal/app/core/contracts/microservices/header-contract"
	"auth/internal/app/core/helpers/errorhandler"
	"auth/internal/app/core/helpers/response"
	"auth/pkg/cache"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthUserMiddleware struct {
	cache cache.Cache
}

func NewAuthUserMiddleware(cache cache.Cache) *AuthUserMiddleware {
	return &AuthUserMiddleware{
		cache: cache,
	}
}

func (m *AuthUserMiddleware) ContextWithAuthUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authUser, err := m.parseAuthUser(ctx)
		if err != nil {
			errMsg := "failed to parse auth user middleware"
			ctx.AbortWithStatusJSON(http.StatusBadRequest, response.ErrorResponse(errMsg))
			errorhandler.FailOnError(err, errMsg)

			return
		}

		auCtx := context.WithValue(ctx.Request.Context(), headercontract.AuthUserKey{}, authUser)
		ctx.Request = ctx.Request.WithContext(auCtx)

		ctx.Next()
	}
}

func (m *AuthUserMiddleware) parseAuthUser(ctx *gin.Context) (headercontract.AuthUser, error) {
	var authUser headercontract.AuthUser

	id, err := strconv.ParseInt(ctx.GetHeader(headercontract.UserIdKey), 10, 64)
	if err != nil {
		return headercontract.AuthUser{}, err
	}

	authUser.ID = id

	email := ctx.GetHeader(headercontract.UserEmailKey)
	if email == "" {
		return headercontract.AuthUser{}, errors.New("user email not found in header")
	}

	authUser.Email = email

	return authUser, nil
}

func (m *AuthUserMiddleware) AuthUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")
		var conflict bool

		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))

			return
		}

		userID, err := m.cache.Get(ctx, fmt.Sprintf("access_token:%s", token))
		if err != nil {
			if errors.Is(err, cache.ErrCacheMiss) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))

				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
			errorhandler.FailOnError(err, "failed to get cached user id")

			return
		}

		email, err := m.cache.Get(ctx, fmt.Sprintf("user_email:%s", string(userID)))
		if err != nil {
			if errors.Is(err, cache.ErrCacheMiss) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))

				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
			errorhandler.FailOnError(err, "failed to get cached user email")

			return
		}

		currentToken, err := m.cache.Get(ctx, fmt.Sprintf("user_id:%s", string(userID)))
		if err != nil {
			if errors.Is(err, cache.ErrCacheMiss) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))

				return
			}

			ctx.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
			errorhandler.FailOnError(err, "failed to get cached access token")

			return
		}

		if string(currentToken) != token {
			conflict = true

			err := m.cache.Delete(ctx, fmt.Sprintf("access_token:%s", token))
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
				errorhandler.FailOnError(err, "failed to delete cached access token")

				return
			}
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.ErrorResponseWithData(
				response.Unauthorized,
				map[string]interface{}{"conflict": conflict},
			))

			return
		}

		ctx.Request.Header.Set(headercontract.UserIdKey, string(userID))
		ctx.Request.Header.Set(headercontract.UserEmailKey, string(email))

		ctx.Next()
	}
}
