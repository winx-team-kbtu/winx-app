package swipe

import (
	"net/http"

	headercontract "winx-match/internal/app/core/contracts/microservices/header-contract"
	"winx-match/internal/app/core/helpers/errorhandler"
	"winx-match/internal/app/core/helpers/response"
	"winx-match/internal/app/core/validation"
	requestdto "winx-match/internal/app/domain/core/dto/requests/swipe"
	matchResource "winx-match/internal/app/domain/resources/match"
	service "winx-match/internal/app/domain/services/general"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.Service
	binder  *validation.Binder
}

func NewHandler(svc service.Service, binder *validation.Binder) *Handler {
	return &Handler{service: svc, binder: binder}
}

// SwipeLeft — POST /swipes/left
func (h *Handler) SwipeLeft(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")
		return
	}

	payload, ok := validation.BindAndValidate[requestdto.SwipeDTO](h.binder, ctx)
	if !ok {
		return
	}

	if err := h.service.HandleSwipeLeft(ctx.Request.Context(), authUser.ID, payload.SwipedID); err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		errorhandler.FailOnError(err, "SwipeLeft service error")
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(nil, response.OK))
}

// SwipeRight — POST /swipes/right
// Returns the created match in the response body when both users swiped right.
func (h *Handler) SwipeRight(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")
		return
	}

	payload, ok := validation.BindAndValidate[requestdto.SwipeDTO](h.binder, ctx)
	if !ok {
		return
	}

	match, created, err := h.service.HandleSwipeRight(ctx.Request.Context(), authUser.ID, payload.SwipedID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		errorhandler.FailOnError(err, "SwipeRight service error")
		return
	}

	if created {
		ctx.JSON(http.StatusOK, response.SuccessResponse(matchResource.NewResource(match), response.OK))
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(nil, response.OK))
}
