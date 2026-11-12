package team

import (
	"errors"
	"net/http"

	headercontract "winx-profile/internal/app/core/contracts/microservices/header-contract"
	"winx-profile/internal/app/core/helpers/errorhandler"
	"winx-profile/internal/app/core/helpers/response"
	"winx-profile/internal/app/core/validation"
	dto "winx-profile/internal/app/domain/core/dto/requests/team"
	resources "winx-profile/internal/app/domain/resources/team"
	service "winx-profile/internal/app/domain/services/team"

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
	case errors.Is(err, service.ErrNotFound):
		return http.StatusNotFound, true
	case errors.Is(err, service.ErrAlreadyInTeam), errors.Is(err, service.ErrTeamFull):
		return http.StatusConflict, true
	case errors.Is(err, service.ErrInvalidName):
		return http.StatusUnprocessableEntity, true
	default:
		return 0, false
	}
}

func fail(ctx *gin.Context, err error, msg string) {
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

	team, err := h.service.Create(ctx.Request.Context(), authUser.ID, payload.CourseID, payload.Name)
	if err != nil {
		fail(ctx, err, "CreateTeam service error")

		return
	}

	ctx.JSON(http.StatusCreated, response.SuccessResponse(resources.NewResource(team), response.Created))
}

func (h *Handler) GetMine(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")

		return
	}

	query, ok := validation.BindAndValidate[dto.MineDTO](h.binder, ctx)
	if !ok {
		return
	}

	team, err := h.service.GetMine(ctx.Request.Context(), authUser.ID, query.CourseID)
	if err != nil {
		fail(ctx, err, "GetMyTeam service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(resources.NewResource(team), response.OK))
}

func (h *Handler) Join(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")

		return
	}

	payload, ok := validation.BindAndValidate[dto.JoinDTO](h.binder, ctx)
	if !ok {
		return
	}

	team, err := h.service.Join(ctx.Request.Context(), authUser.ID, payload.JoinCode)
	if err != nil {
		fail(ctx, err, "JoinTeam service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(resources.NewResource(team), response.Updated))
}

func (h *Handler) Roster(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")

		return
	}

	query, ok := validation.BindAndValidate[dto.MineDTO](h.binder, ctx)
	if !ok {
		return
	}

	roster, err := h.service.Roster(ctx.Request.Context(), authUser.ID, query.CourseID)
	if err != nil {
		fail(ctx, err, "TeamRoster service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(resources.NewRosterResource(roster), response.OK))
}
