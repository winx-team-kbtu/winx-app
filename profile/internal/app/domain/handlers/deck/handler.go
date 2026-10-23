package deck

import (
	"errors"
	"net/http"

	headercontract "winx-profile/internal/app/core/contracts/microservices/header-contract"
	"winx-profile/internal/app/core/helpers/errorhandler"
	"winx-profile/internal/app/core/helpers/response"
	"winx-profile/internal/app/core/validation"
	dto "winx-profile/internal/app/domain/core/dto/requests/deck"
	resources "winx-profile/internal/app/domain/resources/deck"
	service "winx-profile/internal/app/domain/services/deck"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.Service
	binder  *validation.Binder
}

func NewHandler(service service.Service, binder *validation.Binder) *Handler {
	return &Handler{service: service, binder: binder}
}

func (h *Handler) List(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")

		return
	}

	query, ok := validation.BindAndValidate[dto.ListDTO](h.binder, ctx)
	if !ok {
		return
	}

	items, err := h.service.Candidates(ctx.Request.Context(), authUser.ID, query.CourseID, query.Limit, query.Offset)
	if err != nil {
		if errors.Is(err, service.ErrCourseNotSelected) {
			ctx.JSON(http.StatusForbidden, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}
		errorhandler.FailOnError(err, "ListDeckCandidates service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(resources.NewCollection(items), response.OK))
}
