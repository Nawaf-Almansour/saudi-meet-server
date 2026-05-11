package handlers

import (
	"context"
	"time"

	"github.com/Nawaf-Almansour/diwan-server/internal/middleware"
	"github.com/Nawaf-Almansour/diwan-server/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db        *pgxpool.Pool
	jwtSecret string
}

func NewAuthHandler(db *pgxpool.Pool, jwtSecret string) *AuthHandler {
	return &AuthHandler{db: db, jwtSecret: jwtSecret}
}

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email, password, and display_name are required"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash password"})
	}

	var user models.User
	err = h.db.QueryRow(context.Background(),
		`INSERT INTO users (email, password_hash, display_name) VALUES ($1, $2, $3)
		 RETURNING id, email, display_name, avatar_url, status, created_at`,
		req.Email, string(hash), req.DisplayName,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL, &user.Status, &user.CreatedAt)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email already exists"})
	}

	token, err := h.generateToken(&user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate token"})
	}

	return c.Status(fiber.StatusCreated).JSON(AuthResponse{Token: token, User: &user})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	var user models.User
	var passwordHash string
	err := h.db.QueryRow(context.Background(),
		`SELECT id, email, display_name, avatar_url, status, created_at, password_hash FROM users WHERE email = $1`,
		req.Email,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL, &user.Status, &user.CreatedAt, &passwordHash)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	token, err := h.generateToken(&user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate token"})
	}

	return c.JSON(AuthResponse{Token: token, User: &user})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var user models.User
	err := h.db.QueryRow(context.Background(),
		`SELECT id, email, display_name, avatar_url, status, created_at FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL, &user.Status, &user.CreatedAt)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	return c.JSON(user)
}

func (h *AuthHandler) generateToken(user *models.User) (string, error) {
	claims := &middleware.Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}
