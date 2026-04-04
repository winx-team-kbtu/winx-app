package match

import (
	"net/http"
	"strconv"

	headercontract "winx-match/internal/app/core/contracts/microservices/header-contract"
	"winx-match/internal/app/core/helpers/errorhandler"
	"winx-match/internal/app/core/helpers/response"
	"winx-match/internal/app/core/validation"
	servicedto "winx-match/internal/app/domain/core/dto/services/match"
	resources "winx-match/internal/app/domain/resources/match"
	service "winx-match/internal/app/domain/services/match"

	"github.com/gin-gonic/gin"
)

var errStatusMap = map[error]int{
	service.ErrNotFound: http.StatusNotFound,
}

// Handler обрабатывает HTTP-запросы для матчей.
//
// Чтобы добавить новый эндпоинт:
//  1. Добавь метод в Service интерфейс (services/match/service.go)
//  2. Реализуй метод в service struct
//  3. Добавь метод хендлера здесь
//  4. Зарегистрируй роут в api/route.go → initMatchRoutes()
type Handler struct {
	service service.Service
	binder  *validation.Binder
}

func NewHandler(svc service.Service, binder *validation.Binder) *Handler {
	return &Handler{service: svc, binder: binder}
}

// List — GET /matches
// Возвращает список матчей текущего пользователя.
func (h *Handler) List(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	matches, err := h.service.List(ctx.Request.Context(), servicedto.ListDTO{
		UserID: authUser.ID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		errorhandler.FailOnError(err, "List matches service error")
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(resources.NewCollection(matches), response.OK))
}

// Delete — DELETE /matches/:id
// Удаляет матч по id (только если текущий пользователь участник матча).
func (h *Handler) Delete(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse("invalid match id"))
		errorhandler.FailOnError(err, "invalid match id")
		return
	}

	ok, err := h.service.Delete(ctx.Request.Context(), servicedto.DeleteDTO{
		ID:     id,
		UserID: authUser.ID,
	})
	if err != nil {
		if code, found := errStatusMap[err]; found {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}
		errorhandler.FailOnError(err, "Delete match service error")
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(ok, response.Deleted))
}
