package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"winx-chat/internal/app/core/helpers/errorhandler"
	servicedto "winx-chat/internal/app/domain/core/dto/services/message"
	msgResources "winx-chat/internal/app/domain/resources/message"
	msgService "winx-chat/internal/app/domain/services/message"
	"winx-chat/internal/app/models/models"
	"winx-chat/pkg/cache"
	"winx-chat/pkg/graylog/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ---------------------------------------------------------------------------
// Hub — manages all active WebSocket connections, keyed by userID.
// ---------------------------------------------------------------------------

type Hub struct {
	mu      sync.RWMutex
	clients map[int64]*Client // userID → client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[int64]*Client),
	}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// Close old connection if the same user reconnects.
	if old, ok := h.clients[c.userID]; ok {
		old.conn.Close()
	}
	h.clients[c.userID] = c
}

func (h *Hub) Unregister(userID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, userID)
}

func (h *Hub) Send(userID int64, payload []byte) {
	h.mu.RLock()
	c, ok := h.clients[userID]
	h.mu.RUnlock()
	if ok {
		select {
		case c.send <- payload:
		default:
			// buffer full — drop message for this client
		}
	}
}

// ---------------------------------------------------------------------------
// Client — a single WebSocket connection for one authenticated user.
// ---------------------------------------------------------------------------

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID int64
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

// readPump reads incoming messages from the WebSocket connection.
func (c *Client) readPump(ctx context.Context, msgSvc msgService.Service, chatSvc chatFinder) {
	defer func() {
		c.hub.Unregister(c.userID)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logger.Log.Errorf("ws read error user=%d: %v", c.userID, err)
			}
			return
		}

		var incoming incomingMessage
		if err := json.Unmarshal(raw, &incoming); err != nil {
			c.sendError("invalid json")
			continue
		}

		if incoming.ChatID == 0 || strings.TrimSpace(incoming.Content) == "" {
			c.sendError("chat_id and content are required")
			continue
		}

		msg, err := msgSvc.Create(ctx, servicedto.CreateDTO{
			ChatID:   incoming.ChatID,
			SenderID: c.userID,
			Content:  incoming.Content,
		})
		if err != nil {
			if errors.Is(err, msgService.ErrChatNotFound) {
				c.sendError("chat not found")
			} else if errors.Is(err, msgService.ErrNotMember) {
				c.sendError("you are not a member of this chat")
			} else {
				c.sendError("failed to send message")
				errorhandler.FailOnError(err, "ws create message")
			}
			continue
		}

		// Determine the other participant and broadcast.
		chat, err := chatSvc.FindByID(ctx, incoming.ChatID)
		if err != nil {
			continue
		}

		outPayload := buildOutgoing(msg)

		recipientID := chat.UserOneID
		if recipientID == c.userID {
			recipientID = chat.UserTwoID
		}

		// Send to both the sender (ack) and the recipient.
		c.hub.Send(c.userID, outPayload)
		c.hub.Send(recipientID, outPayload)
	}
}

// writePump writes messages from the send channel to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendError(msg string) {
	payload, _ := json.Marshal(outgoingMessage{
		Type:  "error",
		Error: msg,
	})
	select {
	case c.send <- payload:
	default:
	}
}

// ---------------------------------------------------------------------------
// WebSocket messages
// ---------------------------------------------------------------------------

type incomingMessage struct {
	ChatID  int64  `json:"chat_id"`
	Content string `json:"content"`
}

type outgoingMessage struct {
	Type  string                 `json:"type"`
	Data  *msgResources.Resource `json:"data,omitempty"`
	Error string                 `json:"error,omitempty"`
}

func buildOutgoing(msg models.Message) []byte {
	out := outgoingMessage{
		Type: "message",
		Data: msgResources.NewResource(msg),
	}
	b, _ := json.Marshal(out)
	return b
}

// chatFinder is a minimal interface so ws.go doesn't depend on the full chat service.
type chatFinder interface {
	FindByID(ctx context.Context, id int64) (models.Chat, error)
}

// ---------------------------------------------------------------------------
// WebSocket upgrade handler
// ---------------------------------------------------------------------------

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// handleWS upgrades the HTTP connection to WebSocket.
// Auth is done via query param: ?token=<access_token>
func (s *Server) handleWS(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "token query param is required"})
		return
	}

	userID, err := s.authenticateToken(ctx, token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.Log.Errorf("ws upgrade: %v", err)
		return
	}

	client := &Client{
		hub:    s.hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
	}

	s.hub.Register(client)

	go client.writePump()
	go client.readPump(context.Background(), s.msgService, s.chatService)
}

// authenticateToken validates the access token via Redis cache (same logic as auth middleware).
func (s *Server) authenticateToken(ctx context.Context, token string) (int64, error) {
	userIDBytes, err := s.cache.Get(ctx, fmt.Sprintf("access_token:%s", token))
	if err != nil {
		if errors.Is(err, cache.ErrCacheMiss) {
			return 0, fmt.Errorf("token not found")
		}
		return 0, err
	}

	// Verify the token is the current active token for this user.
	currentToken, err := s.cache.Get(ctx, fmt.Sprintf("user_id:%s", string(userIDBytes)))
	if err != nil {
		return 0, err
	}
	if string(currentToken) != token {
		return 0, fmt.Errorf("token conflict")
	}

	var uid int64
	if _, err := fmt.Sscanf(string(userIDBytes), "%d", &uid); err != nil {
		return 0, fmt.Errorf("parse user id: %w", err)
	}

	return uid, nil
}
