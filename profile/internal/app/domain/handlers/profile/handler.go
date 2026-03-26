package profile

import (
	headercontract "winx-profile/internal/app/core/contracts/microservices/header-contract"
	"winx-profile/internal/app/core/helpers/errorhandler"
	"winx-profile/internal/app/core/helpers/response"
	"winx-profile/internal/app/core/validation"
	reqDto "winx-profile/internal/app/domain/core/dto/requests/profile"
	serviceDto "winx-profile/internal/app/domain/core/dto/services/profile"
	profileResource "winx-profile/internal/app/domain/resources/profile"
	service "winx-profile/internal/app/domain/services/profile"
	"net/http"

	"github.com/gin-gonic/gin"
)

var errStatusMap = map[error]int{
	service.ErrNotFound: http.StatusNotFound,
}

type Handler struct {
	service service.Service
	binder  *validation.Binder
}

func NewHandler(svc service.Service, binder *validation.Binder) *Handler {
	return &Handler{service: svc, binder: binder}
}

func (h *Handler) Get(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		return
	}

	profile, err := h.service.Get(ctx.Request.Context(), authUser.ID)
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}
		errorhandler.FailOnError(err, "Get profile service error")
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(profileResource.NewResource(profile), response.OK))
}

func (h *Handler) Update(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		return
	}

	payload, ok := validation.BindAndValidate[reqDto.UpdateDTO](h.binder, ctx)
	if !ok {
		return
	}

	profile, err := h.service.Update(ctx.Request.Context(), serviceDto.UpdateDTO{
		UserID:    authUser.ID,
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Bio:       payload.Bio,
		AvatarURL: payload.AvatarURL,
	})
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}
		errorhandler.FailOnError(err, "Update profile service error")
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(profileResource.NewResource(profile), response.Updated))
}

func (h *Handler) UpdateRole(ctx *gin.Context) {
	payload, ok := validation.BindAndValidate[reqDto.UpdateRoleDTO](h.binder, ctx)
	if !ok {
		return
	}

	profile, err := h.service.UpdateRole(ctx.Request.Context(), serviceDto.UpdateRoleDTO{
		UserID: payload.UserID,
		Role:   payload.Role,
	})
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}
		errorhandler.FailOnError(err, "UpdateRole profile service error")
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(profileResource.NewResource(profile), response.Updated))
}
