package api

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"

	db "github.com/Maestrominds/lcdp-restaurant/internal/sqlc"
)

type userRequest struct {
	Name   string `json:"name"`
	Role   string `json:"role"`
	Phone  string `json:"phone"`
}

// Using isValidUserRole from auth.go

func (h *Handler) CreateUser(c *fiber.Ctx) error {
	var req userRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Role = strings.ToLower(strings.TrimSpace(req.Role))
	req.Phone = strings.TrimSpace(req.Phone)

	if req.Name == "" || req.Role == "" || req.Phone == "" {
		return writeError(c, fiber.StatusBadRequest, "missing_fields", "name, role, and phone are required")
	}

	if !isValidUserRole(req.Role) {
		return writeError(c, fiber.StatusBadRequest, "invalid_role", "role must be waiter, manager, or admin")
	}

	user, err := h.DB.CreateUser(context.Background(), db.CreateUserParams{
		Name:  req.Name,
		Role:  db.UserRole(req.Role),
		Phone: req.Phone,
	})
	if err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

func (h *Handler) ListUsers(c *fiber.Ctx) error {
	users, err := h.DB.ListUsers(context.Background())
	if err != nil {
		return handleDBError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(users)
}

func (h *Handler) GetUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	user, err := h.DB.GetUser(context.Background(), id)
	if err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	var req userRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Role = strings.ToLower(strings.TrimSpace(req.Role))
	req.Phone = strings.TrimSpace(req.Phone)
	if req.Name == "" || req.Role == "" || req.Phone == "" {
		return writeError(c, fiber.StatusBadRequest, "missing_fields", "name, role, and phone are required")
	}

	if !isValidUserRole(req.Role) {
		return writeError(c, fiber.StatusBadRequest, "invalid_role", "role must be waiter, manager, or admin")
	}

	user, err := h.DB.UpdateUser(context.Background(), db.UpdateUserParams{
		ID:     id,
		Name:   req.Name,
		Role:   db.UserRole(req.Role),
		Phone:  req.Phone,
	})
	if err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	if err := h.DB.DeleteUser(context.Background(), id); err != nil {
		return handleDBError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListWaiters(c *fiber.Ctx) error {
	users, err := h.DB.ListUsers(context.Background())
	if err != nil {
		return handleDBError(c, err)
	}

	waiters := make([]db.User, 0)
	for _, u := range users {
		if u.Role == "waiter" {
			waiters = append(waiters, u)
		}
	}
	return c.Status(fiber.StatusOK).JSON(waiters)
}
