package recommendation

import (
	"net/http"

	headercontract "winx-recommendation/internal/app/core/contracts/microservices/header-contract"
	"winx-recommendation/internal/app/core/helpers/errorhandler"
	"winx-recommendation/internal/app/core/helpers/response"
	"winx-recommendation/internal/app/core/validation"
	requestdto "winx-recommendation/internal/app/domain/core/dto/requests/recommendation"
	svcDto "winx-recommendation/internal/app/domain/core/dto/services/recommendation"
	resources "winx-recommendation/internal/app/domain/resources/recommendation"
	service "winx-recommendation/internal/app/domain/services/recommendation"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.Service
	binder  *validation.Binder
}

func NewHandler(svc service.Service, binder *validation.Binder) *Handler {
	return &Handler{service: svc, binder: binder}
}

// List — GET /recommendations
// Returns a ranked list of recommended profiles for the authenticated user.
// Query params: limit (default 20, max 100), offset (default 0).
func (h *Handler) List(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")
		return
	}

	payload, ok := validation.BindAndValidate[requestdto.ListDTO](h.binder, ctx)
	if !ok {
		return
	}

	candidates, err := h.service.List(ctx.Request.Context(), svcDto.ListDTO{
		UserID: authUser.ID,
		Limit:  payload.Limit,
		Offset: payload.Offset,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		errorhandler.FailOnError(err, "recommendations list error")
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(resources.NewCollection(candidates), response.OK))
}
