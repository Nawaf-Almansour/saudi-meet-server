package handlers

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Nawaf-Almansour/diwan-server/internal/models"
)

type WorkspaceHandler struct {
	db *pgxpool.Pool
}

func NewWorkspaceHandler(db *pgxpool.Pool) *WorkspaceHandler {
	return &WorkspaceHandler{db: db}
}

type CreateWorkspaceRequest struct {
	Name string `json:"name"`
}

func (h *WorkspaceHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req CreateWorkspaceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	slug := strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))

	tx, err := h.db.Begin(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to start transaction"})
	}
	defer tx.Rollback(context.Background())

	var ws models.Workspace
	err = tx.QueryRow(context.Background(),
		`INSERT INTO workspaces (name, slug, owner_id) VALUES ($1, $2, $3)
		 RETURNING id, name, slug, owner_id, created_at`,
		req.Name, slug, userID,
	).Scan(&ws.ID, &ws.Name, &ws.Slug, &ws.OwnerID, &ws.CreatedAt)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "workspace slug already exists"})
	}

	// Add owner as admin member
	_, err = tx.Exec(context.Background(),
		`INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'admin')`,
		ws.ID, userID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to add member"})
	}

	// Create default #general channel
	_, err = tx.Exec(context.Background(),
		`INSERT INTO channels (workspace_id, name, slug, description, visibility, created_by) VALUES ($1, 'عام', 'general', 'القناة العامة للمحادثات', 'public', $2)`,
		ws.ID, userID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create default channel"})
	}

	if err := tx.Commit(context.Background()); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to commit"})
	}

	return c.Status(fiber.StatusCreated).JSON(ws)
}

func (h *WorkspaceHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	rows, err := h.db.Query(context.Background(),
		`SELECT w.id, w.name, w.slug, w.owner_id, w.created_at
		 FROM workspaces w
		 JOIN workspace_members wm ON w.id = wm.workspace_id
		 WHERE wm.user_id = $1
		 ORDER BY w.created_at DESC`,
		userID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch workspaces"})
	}
	defer rows.Close()

	var workspaces []models.Workspace
	for rows.Next() {
		var ws models.Workspace
		if err := rows.Scan(&ws.ID, &ws.Name, &ws.Slug, &ws.OwnerID, &ws.CreatedAt); err != nil {
			continue
		}
		workspaces = append(workspaces, ws)
	}

	if workspaces == nil {
		workspaces = []models.Workspace{}
	}

	return c.JSON(workspaces)
}

func (h *WorkspaceHandler) Get(c *fiber.Ctx) error {
	wsID := c.Params("id")
	userID := c.Locals("user_id").(string)

	// Verify membership
	var count int
	err := h.db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`,
		wsID, userID,
	).Scan(&count)
	if err != nil || count == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "not a member of this workspace"})
	}

	var ws models.Workspace
	err = h.db.QueryRow(context.Background(),
		`SELECT id, name, slug, owner_id, created_at FROM workspaces WHERE id = $1`,
		wsID,
	).Scan(&ws.ID, &ws.Name, &ws.Slug, &ws.OwnerID, &ws.CreatedAt)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "workspace not found"})
	}

	return c.JSON(ws)
}

// Join workspace by slug
func (h *WorkspaceHandler) Join(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req struct {
		Slug string `json:"slug"`
	}
	if err := c.BodyParser(&req); err != nil || req.Slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "slug is required"})
	}

	// Find workspace by slug
	var ws models.Workspace
	err := h.db.QueryRow(context.Background(),
		`SELECT id, name, slug, owner_id, created_at FROM workspaces WHERE slug = $1`,
		req.Slug,
	).Scan(&ws.ID, &ws.Name, &ws.Slug, &ws.OwnerID, &ws.CreatedAt)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "workspace not found"})
	}

	// Add user as member
	_, err = h.db.Exec(context.Background(),
		`INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'member') ON CONFLICT DO NOTHING`,
		ws.ID, userID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to join workspace"})
	}

	// Auto-join all public channels
	rows, err := h.db.Query(context.Background(),
		`SELECT id FROM channels WHERE workspace_id = $1 AND visibility = 'public' AND archived_at IS NULL`,
		ws.ID,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var chID string
			if rows.Scan(&chID) == nil {
				h.db.Exec(context.Background(),
					`INSERT INTO channel_members (channel_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
					chID, userID,
				)
			}
		}
	}

	return c.JSON(fiber.Map{"status": "joined", "workspace": ws})
}

// List all available workspaces (public discovery)
func (h *WorkspaceHandler) ListAll(c *fiber.Ctx) error {
	rows, err := h.db.Query(context.Background(),
		`SELECT id, name, slug, owner_id, created_at FROM workspaces ORDER BY created_at DESC LIMIT 50`,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch workspaces"})
	}
	defer rows.Close()

	var workspaces []models.Workspace
	for rows.Next() {
		var ws models.Workspace
		if err := rows.Scan(&ws.ID, &ws.Name, &ws.Slug, &ws.OwnerID, &ws.CreatedAt); err != nil {
			continue
		}
		workspaces = append(workspaces, ws)
	}

	if workspaces == nil {
		workspaces = []models.Workspace{}
	}

	return c.JSON(workspaces)
}
