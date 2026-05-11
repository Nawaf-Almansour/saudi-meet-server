package handlers

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/Nawaf-Almansour/diwan-server/internal/middleware"
	"github.com/Nawaf-Almansour/diwan-server/internal/ws"
	"github.com/gofiber/contrib/websocket"
	"github.com/golang-jwt/jwt/v5"
)

type WSHandler struct {
	hub       *ws.Hub
	jwtSecret string
}

func NewWSHandler(hub *ws.Hub, jwtSecret string) *WSHandler {
	return &WSHandler{hub: hub, jwtSecret: jwtSecret}
}

type WSMessage struct {
	Action    string `json:"action"`
	ChannelID string `json:"channel_id,omitempty"`
}

func (h *WSHandler) Handle(conn *websocket.Conn) {
	// Authenticate via query param token
	tokenStr := conn.Query("token")
	if tokenStr == "" {
		conn.WriteMessage(websocket.CloseMessage, []byte("missing token"))
		conn.Close()
		return
	}

	token, err := jwt.ParseWithClaims(tokenStr, &middleware.Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		conn.WriteMessage(websocket.CloseMessage, []byte("invalid token"))
		conn.Close()
		return
	}

	claims := token.Claims.(*middleware.Claims)

	client := &ws.Client{
		Conn:     conn,
		UserID:   claims.UserID,
		Channels: make(map[string]bool),
	}

	h.hub.Register(client)
	defer h.hub.Unregister(client)

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}

		switch strings.ToLower(msg.Action) {
		case "subscribe":
			if msg.ChannelID != "" {
				h.hub.SubscribeToChannel(client, msg.ChannelID)
			}
		case "unsubscribe":
			if msg.ChannelID != "" {
				h.hub.UnsubscribeFromChannel(client, msg.ChannelID)
			}
		default:
			log.Printf("unknown ws action: %s", msg.Action)
		}
	}
}
