package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/pkg/dtos"
)

type Handler struct {
	service  *AuthService
	sessions SessionStore
	secure   bool
}

func NewHandler(service *AuthService, sessions SessionStore, secure bool) *Handler {
	return &Handler{
		service:  service,
		sessions: sessions,
		secure:   secure,
	}
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	user, err := h.service.Register(c.Context(), RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if err != nil {
		return common.HandlerError(c, err)
	}

	err = h.createSessionAndSetCookie(c, user)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(mapToUserDto(user))
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	user, err := h.service.Login(c.Context(), LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return common.HandlerError(c, err)
	}

	err = h.createSessionAndSetCookie(c, user)
	if err != nil {
		return err
	}

	return c.JSON(mapToUserDto(user))
}

func (h *Handler) createSessionAndSetCookie(c *fiber.Ctx, user User) error {
	sessionID, err := h.sessions.Create(c.Context(), user.ID)
	if err != nil {
		return common.HandlerError(c, err)
	}
	c.Cookie(&fiber.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  time.Now().Add(sessionTTL),
		Secure:   h.secure,
		HTTPOnly: true,
		SameSite: "Strict",
		MaxAge:   int(sessionTTL.Seconds()),
	})
	return nil
}

func mapToUserDto(user User) dtos.UserDto {
	return dtos.UserDto{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
