package chat

import (
	"net/http"
	"strconv"

	headercontract "winx-chat/internal/app/core/contracts/microservices/header-contract"
	"winx-chat/internal/app/core/helpers/errorhandler"
	"winx-chat/internal/app/core/helpers/response"
	"winx-chat/internal/app/core/validation"
	chatdto "winx-chat/internal/app/domain/core/dto/services/chat"
	msgdto "winx-chat/internal/app/domain/core/dto/services/message"
	chatResources "winx-chat/internal/app/domain/resources/chat"
	msgResources "winx-chat/internal/app/domain/resources/message"
	chatService "winx-chat/internal/app/domain/services/chat"
	msgService "winx-chat/internal/app/domain/services/message"

	"github.com/gin-gonic/gin"
)

var errStatusMap = map[error]int{
	chatService.ErrNotFound:    http.StatusNotFound,
	msgService.ErrChatNotFound: http.StatusNotFound,
	msgService.ErrNotMember:    http.StatusForbidden,
}

type Handler struct {
	chatService chatService.Service
	msgService  msgService.Service
	binder      *validation.Binder
}

func NewHandler(cs chatService.Service, ms msgService.Service, binder *validation.Binder) *Handler {
	return &Handler{chatService: cs, msgService: ms, binder: binder}
}

// List — GET /chats
// Возвращает список чатов текущего пользователя.
func (h *Handler) List(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	chats, err := h.chatService.List(ctx.Request.Context(), chatdto.ListDTO{
		UserID: authUser.ID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		errorhandler.FailOnError(err, "List chats service error")
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(chatResources.NewCollection(chats), response.OK))
}

// Messages — GET /chats/:id/messages
// Возвращает историю сообщений для указанного чата.
func (h *Handler) Messages(ctx *gin.Context) {
	authUser, err := headercontract.GetAuthUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(response.Unauthorized))
		errorhandler.FailOnError(err, "failed to get auth user")
		return
	}

	chatID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse("invalid chat id"))
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	messages, err := h.msgService.List(ctx.Request.Context(), msgdto.ListDTO{
		ChatID: chatID,
		UserID: authUser.ID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		if code, found := errStatusMap[err]; found {
			ctx.JSON(code, response.ErrorResponse(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(response.ServerError))
		}
		errorhandler.FailOnError(err, "Messages service error")
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse(msgResources.NewCollection(messages), response.OK))
}
