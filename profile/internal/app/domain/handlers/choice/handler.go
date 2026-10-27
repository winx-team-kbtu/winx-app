package choice

import (
	"errors"
	"net/http"

	headercontract "winx-profile/internal/app/core/contracts/microservices/header-contract"
	"winx-profile/internal/app/core/helpers/errorhandler"
	"winx-profile/internal/app/core/helpers/response"
	"winx-profile/internal/app/core/validation"
	dto "winx-profile/internal/app/domain/core/dto/requests/choice"
	deckResources "winx-profile/internal/app/domain/resources/deck"
	service "winx-profile/internal/app/domain/services/choice"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.Service
	binder  *validation.Binder
}

func NewHandler(service service.Service, binder *validation.Binder) *Handler {
	return &Handler{service: service, binder: binder}
}

func statusFor(err error) (int, bool) {
	switch {
	case errors.Is(err, service.ErrCourseNotSelected):
		return http.StatusForbidden, true
	case errors.Is(err, service.ErrInvalidCandidate), errors.Is(err, service.ErrInvalidChoice):
		return http.StatusUnprocessableEntity, true
	default:
		return 0, false
	}
}

func (h *Handler) fail(ctx *gin.Context, err error, msg string) {
	if code, ok := statusFor(err); ok {
		ctx.JSON(code, response.ErrorResponse(err.Error()))
	} else {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
	}

	errorhandler.FailOnError(err, msg)
}

func (h *Handler) Store(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")

		return
	}

	payload, ok := validation.BindAndValidate[dto.StoreDTO](h.binder, ctx)
	if !ok {
		return
	}

	if err := h.service.Choose(ctx.Request.Context(), authUser.ID, payload.CandidateID, payload.CourseID, payload.Choice); err != nil {
		h.fail(ctx, err, "StoreChoice service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(nil, response.Updated))
}

func (h *Handler) Shortlist(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")

		return
	}

	query, ok := validation.BindAndValidate[dto.ShortlistDTO](h.binder, ctx)
	if !ok {
		return
	}

	items, err := h.service.Shortlist(ctx.Request.Context(), authUser.ID, query.CourseID, query.Limit, query.Offset)
	if err != nil {
		h.fail(ctx, err, "Shortlist service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(deckResources.NewCollection(items), response.OK))
}
