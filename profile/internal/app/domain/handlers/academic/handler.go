package academic

import (
	"errors"
	"net/http"

	headercontract "winx-profile/internal/app/core/contracts/microservices/header-contract"
	"winx-profile/internal/app/core/helpers/errorhandler"
	"winx-profile/internal/app/core/helpers/response"
	"winx-profile/internal/app/core/validation"
	dto "winx-profile/internal/app/domain/core/dto/requests/academic"
	resources "winx-profile/internal/app/domain/resources/academic"
	service "winx-profile/internal/app/domain/services/academic"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.Service
	binder  *validation.Binder
}

func NewHandler(service service.Service, binder *validation.Binder) *Handler {
	return &Handler{service: service, binder: binder}
}

func (h *Handler) GetMe(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")

		return
	}

	profile, err := h.service.Get(ctx.Request.Context(), authUser.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		errorhandler.FailOnError(err, "GetAcademicProfile service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(resources.NewResource(profile), response.OK))
}

func (h *Handler) SaveMe(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")

		return
	}

	payload, ok := validation.BindAndValidate[dto.SaveDTO](h.binder, ctx)
	if !ok {
		return
	}

	profile, err := h.service.Save(ctx.Request.Context(), authUser.ID, service.SaveInput{
		Skills:        payload.Skills,
		PreferredRole: payload.PreferredRole,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			ctx.JSON(http.StatusUnprocessableEntity, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}
		errorhandler.FailOnError(err, "SaveAcademicProfile service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(resources.NewResource(profile), response.Updated))
}
