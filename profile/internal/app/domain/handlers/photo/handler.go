package photo

import (
	"errors"
	"net/http"

	headercontract "winx-profile/internal/app/core/contracts/microservices/header-contract"
	"winx-profile/internal/app/core/helpers/errorhandler"
	"winx-profile/internal/app/core/helpers/response"
	serviceDto "winx-profile/internal/app/domain/core/dto/services/photo"
	resources "winx-profile/internal/app/domain/resources/photo"
	service "winx-profile/internal/app/domain/services/photo"

	"github.com/gin-gonic/gin"
)

var errStatusMap = map[error]int{
	service.ErrNotFound: http.StatusNotFound,
}

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetMe(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")

		return
	}

	photo, err := h.service.GetByUserID(ctx.Request.Context(), authUser.ID)
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}
		errorhandler.FailOnError(err, "GetPhoto service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(resources.NewResource(photo), response.OK))
}

func (h *Handler) Store(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")

		return
	}

	fileHeader, err := ctx.FormFile("photo")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse("photo file is required"))
		errorhandler.FailOnError(err, "failed to read photo file")

		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse("failed to open photo file"))
		errorhandler.FailOnError(err, "failed to open photo file")

		return
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			errorhandler.FailOnError(cerr, "failed to close photo file")
		}
	}()

	photo, created, err := h.service.Store(ctx.Request.Context(), serviceDto.StoreDTO{
		UserID:      authUser.ID,
		Email:       authUser.Email,
		File:        file,
		FileName:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		SizeBytes:   fileHeader.Size,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidPhoto) {
			ctx.JSON(http.StatusBadRequest, response.ErrorResponse(err.Error()))
			errorhandler.FailOnError(err, "StorePhoto invalid photo")

			return
		}

		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		errorhandler.FailOnError(err, "StorePhoto service error")

		return
	}

	status := http.StatusOK
	message := response.Updated
	if created {
		status = http.StatusCreated
		message = response.Created
	}

	ctx.JSON(status, response.SuccessResponse(resources.NewResource(photo), message))
}
