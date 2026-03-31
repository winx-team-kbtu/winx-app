package profile

import (
	"net/http"

	headercontract "winx-profile/internal/app/core/contracts/microservices/header-contract"
	"winx-profile/internal/app/core/helpers/errorhandler"
	"winx-profile/internal/app/core/helpers/response"
	"winx-profile/internal/app/core/validation"
	dto "winx-profile/internal/app/domain/core/dto/requests/profile"
	serviceDto "winx-profile/internal/app/domain/core/dto/services/profile"
	resources "winx-profile/internal/app/domain/resources/profile"
	service "winx-profile/internal/app/domain/services/profile"

	"github.com/gin-gonic/gin"
)

var errStatusMap = map[error]int{
	service.ErrNotFound:           http.StatusNotFound,
	service.ErrInvalidInterestIDs: http.StatusUnprocessableEntity,
}

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

	details, err := h.service.GetDetailsByUserID(ctx.Request.Context(), authUser.ID)
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}
		errorhandler.FailOnError(err, "GetProfile service error")

		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(resources.NewResource(details.Profile, details.Interests), response.OK))
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

	var location *serviceDto.LocationDTO
	if payload.Location != nil {
		location = &serviceDto.LocationDTO{
			City:    payload.Location.City,
			Country: payload.Location.Country,
		}
		if payload.Location.CurrentLocation != nil {
			location.CurrentLocation = &serviceDto.CurrentLocationDTO{
				Latitude:  payload.Location.CurrentLocation.Latitude,
				Longitude: payload.Location.CurrentLocation.Longitude,
			}
		}
	}

	profile, created, err := h.service.Store(ctx.Request.Context(), serviceDto.StoreDTO{
		UserID:       authUser.ID,
		Email:        authUser.Email,
		FirstName:    payload.FirstName,
		BirthDate:    payload.BirthDate,
		Gender:       payload.Gender,
		ShowGender:   payload.ShowGender,
		ShowAge:      payload.ShowAge,
		AboutMe:      payload.AboutMe,
		JobTitle:     payload.JobTitle,
		Company:      payload.Company,
		School:       payload.School,
		InterestIDs:  payload.InterestIDs,
		InterestedIn: payload.InterestedIn,
		LookingFor:   payload.LookingFor,
		Lifestyle:    payload.Lifestyle,
		Preferences:  payload.Preferences,
		Location:     location,
	})
	if err != nil {
		if code, ok := errStatusMap[err]; ok {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}
		errorhandler.FailOnError(err, "StoreProfile service error")

		return
	}

	status := http.StatusOK
	message := response.Updated
	if created {
		status = http.StatusCreated
		message = response.Created
	}

	ctx.JSON(status, response.SuccessResponse(resources.NewResource(profile, nil), message))
}
