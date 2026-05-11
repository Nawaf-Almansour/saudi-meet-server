package handlers

import (
	"context"
	"strings"

	"github.com/Nawaf-Almansour/diwan-server/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChannelHandler struct {
	db *pgxpool.Pool
}

func NewChannelHandler(db *pgxpool.Pool) *ChannelHandler {
	return &ChannelHandler{db: db}
}

type CreateChannelRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
}

func (h *ChannelHandler) Create(c *fiber.Ctx) error {
	wsID := c.Params("workspaceId")
	userID := c.Locals("user_id").(string)

	var req CreateChannelRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	if req.Visibility == "" {
		req.Visibility = "public"
	}

	slug := strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))

	tx, err := h.db.Begin(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to start transaction"})
	}
	defer tx.Rollback(context.Background())

	var ch models.Channel
	err = tx.QueryRow(context.Background(),
		`INSERT INTO channels (workspace_id, name, slug, description, visibility, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, workspace_id, name, slug, description, visibility, created_by, created_at`,
		wsID, req.Name, slug, req.Description, req.Visibility, userID,
	).Scan(&ch.ID, &ch.WorkspaceID, &ch.Name, &ch.Slug, &ch.Description, &ch.Visibility, &ch.CreatedBy, &ch.CreatedAt)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "channel slug already exists in this workspace"})
	}

	// Auto-join the creator
	_, err = tx.Exec(context.Background(),
		`INSERT INTO channel_members (channel_id, user_id, role) VALUES ($1, $2, 'admin')`,
		ch.ID, userID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to add member"})
	}

	if err := tx.Commit(context.Background()); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to commit"})
	}

	return c.Status(fiber.StatusCreated).JSON(ch)
}

func (h *ChannelHandler) List(c *fiber.Ctx) error {
	wsID := c.Params("workspaceId")

	rows, err := h.db.Query(context.Background(),
		`SELECT id, workspace_id, name, slug, description, visibility, created_by, created_at
		 FROM channels
		 WHERE workspace_id = $1 AND archived_at IS NULL
		 ORDER BY created_at ASC`,
		wsID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch channels"})
	}
	defer rows.Close()

	var channels []models.Channel
	for rows.Next() {
		var ch models.Channel
		if err := rows.Scan(&ch.ID, &ch.WorkspaceID, &ch.Name, &ch.Slug, &ch.Description, &ch.Visibility, &ch.CreatedBy, &ch.CreatedAt); err != nil {
			continue
		}
		channels = append(channels, ch)
	}

	if channels == nil {
		channels = []models.Channel{}
	}

	return c.JSON(channels)
}

func (h *ChannelHandler) Join(c *fiber.Ctx) error {
	channelID := c.Params("id")
	userID := c.Locals("user_id").(string)

	_, err := h.db.Exec(context.Background(),
		`INSERT INTO channel_members (channel_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		channelID, userID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to join channel"})
	}

	return c.JSON(fiber.Map{"status": "joined"})
}
