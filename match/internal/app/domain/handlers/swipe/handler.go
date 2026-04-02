package swipe

import (
	"net/http"

	headercontract "winx-match/internal/app/core/contracts/microservices/header-contract"
	"winx-match/internal/app/core/helpers/errorhandler"
	"winx-match/internal/app/core/helpers/response"
	"winx-match/internal/app/core/validation"
	requestdto "winx-match/internal/app/domain/core/dto/requests/swipe"
	resources "winx-match/internal/app/domain/resources/swipe"
	service "winx-match/internal/app/domain/services/swipe"

	"github.com/gin-gonic/gin"
)

// Handler обрабатывает HTTP-запросы для свайпов.
//
// Чтобы добавить новый эндпоинт:
//  1. Добавь метод в Service интерфейс (services/swipe/service.go)
//  2. Реализуй метод в service struct
//  3. Добавь метод хендлера здесь
//  4. Зарегистрируй роут в api/route.go → initSwipeRoutes()
type Handler struct {
	service service.Service
	binder  *validation.Binder
}

func NewHandler(svc service.Service, binder *validation.Binder) *Handler {
	return &Handler{service: svc, binder: binder}
}

// SwipeLeft — POST /swipes/left
// Записывает свайп влево текущего пользователя.
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

	swipe, err := h.service.Create(ctx.Request.Context(), authUser.ID, payload.SwipedID, "left")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		errorhandler.FailOnError(err, "SwipeLeft service error")
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(resources.NewResource(swipe), response.OK))
}

// SwipeRight — POST /swipes/right
// Записывает свайп вправо и инициирует проверку матча через Kafka (handleSwipeRight в server.go).
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

	swipe, err := h.service.Create(ctx.Request.Context(), authUser.ID, payload.SwipedID, "right")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		errorhandler.FailOnError(err, "SwipeRight service error")
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(resources.NewResource(swipe), response.OK))
}
