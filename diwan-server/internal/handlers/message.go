package handlers

import (
	"context"
	"strconv"

	"github.com/Nawaf-Almansour/diwan-server/internal/models"
	"github.com/Nawaf-Almansour/diwan-server/internal/ws"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageHandler struct {
	db  *pgxpool.Pool
	hub *ws.Hub
}

func NewMessageHandler(db *pgxpool.Pool, hub *ws.Hub) *MessageHandler {
	return &MessageHandler{db: db, hub: hub}
}

type SendMessageRequest struct {
	Body     string  `json:"body"`
	ParentID *string `json:"parent_id,omitempty"`
}

func (h *MessageHandler) Send(c *fiber.Ctx) error {
	channelID := c.Params("id")
	userID := c.Locals("user_id").(string)

	var req SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Body == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "body is required"})
	}

	var msg models.Message
	err := h.db.QueryRow(context.Background(),
		`INSERT INTO messages (channel_id, sender_id, parent_id, body, message_type)
		 VALUES ($1, $2, $3, $4, 'text')
		 RETURNING id, channel_id, sender_id, parent_id, body, message_type, created_at, updated_at`,
		channelID, userID, req.ParentID, req.Body,
	).Scan(&msg.ID, &msg.ChannelID, &msg.SenderID, &msg.ParentID, &msg.Body, &msg.MessageType, &msg.CreatedAt, &msg.UpdatedAt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to send message"})
	}

	// Get sender name for broadcast
	var senderName string
	_ = h.db.QueryRow(context.Background(),
		`SELECT display_name FROM users WHERE id = $1`, userID,
	).Scan(&senderName)
	msg.SenderName = senderName

	// Broadcast via WebSocket
	h.hub.BroadcastToChannel(channelID, ws.Event{
		Type:    "message.new",
		Payload: msg,
	})

	return c.Status(fiber.StatusCreated).JSON(msg)
}

func (h *MessageHandler) List(c *fiber.Ctx) error {
	channelID := c.Params("id")
	limitStr := c.Query("limit", "50")
	offsetStr := c.Query("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := h.db.Query(context.Background(),
		`SELECT m.id, m.channel_id, m.sender_id, u.display_name, m.parent_id, m.body, m.message_type, m.created_at, m.updated_at
		 FROM messages m
		 JOIN users u ON m.sender_id = u.id
		 WHERE m.channel_id = $1 AND m.deleted_at IS NULL
		 ORDER BY m.created_at ASC
		 LIMIT $2 OFFSET $3`,
		channelID, limit, offset,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch messages"})
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(&msg.ID, &msg.ChannelID, &msg.SenderID, &msg.SenderName, &msg.ParentID, &msg.Body, &msg.MessageType, &msg.CreatedAt, &msg.UpdatedAt); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	if messages == nil {
		messages = []models.Message{}
	}

	return c.JSON(messages)
}
