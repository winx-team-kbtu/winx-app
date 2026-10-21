package course

import (
	"net/http"

	headercontract "winx-profile/internal/app/core/contracts/microservices/header-contract"
	"winx-profile/internal/app/core/helpers/errorhandler"
	"winx-profile/internal/app/core/helpers/response"
	"winx-profile/internal/app/core/validation"
	dto "winx-profile/internal/app/domain/core/dto/requests/course"
	resources "winx-profile/internal/app/domain/resources/course"
	service "winx-profile/internal/app/domain/services/course"

	"github.com/gin-gonic/gin"
)

var errStatusMap = map[error]int{
	service.ErrInvalidCourseIDs: http.StatusUnprocessableEntity,
}

type Handler struct {
	service service.Service
	binder  *validation.Binder
}

func NewHandler(service service.Service, binder *validation.Binder) *Handler {
	return &Handler{service: service, binder: binder}
}

func (h *Handler) List(ctx *gin.Context) {
	items, err := h.service.List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		errorhandler.FailOnError(err, "ListCourses service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(resources.NewCollection(items), response.OK))
}

func (h *Handler) GetMine(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")

		return
	}

	items, err := h.service.GetUserCourses(ctx.Request.Context(), authUser.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		errorhandler.FailOnError(err, "GetUserCourses service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(resources.NewCollection(items), response.OK))
}

func (h *Handler) SaveMine(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")

		return
	}

	payload, ok := validation.BindAndValidate[dto.SaveUserCoursesDTO](h.binder, ctx)
	if !ok {
		return
	}

	items, err := h.service.SaveUserCourses(ctx.Request.Context(), authUser.ID, payload.CourseIDs)
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}
		errorhandler.FailOnError(err, "SaveUserCourses service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(resources.NewCollection(items), response.Updated))
}
