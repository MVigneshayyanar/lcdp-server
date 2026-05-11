package api

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"

	db "github.com/Maestrominds/lcdp-restaurant/internal/sqlc"
)

type vendorRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
}

func (h *Handler) CreateVendor(c *fiber.Ctx) error {
	var req vendorRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Address = strings.TrimSpace(req.Address)
	req.Email = strings.TrimSpace(req.Email)
	req.Phone = strings.TrimSpace(req.Phone)

	if req.Name == "" || req.Address == "" || req.Email == "" || req.Phone == "" {
		return writeError(c, fiber.StatusBadRequest, "missing_fields", "name, address, email, and phone are required")
	}

	vendor, err := h.DB.CreateVendor(context.Background(), db.CreateVendorParams{
		Name:    req.Name,
		Address: req.Address,
		Email:   req.Email,
		Phone:   req.Phone,
	})
	if err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(vendor)
}

func (h *Handler) ListVendors(c *fiber.Ctx) error {
	vendors, err := h.DB.ListVendors(context.Background())
	if err != nil {
		return handleDBError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(vendors)
}

func (h *Handler) GetVendor(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	vendor, err := h.DB.GetVendor(context.Background(), id)
	if err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(vendor)
}

func (h *Handler) UpdateVendor(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	var req vendorRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Address = strings.TrimSpace(req.Address)
	req.Email = strings.TrimSpace(req.Email)
	req.Phone = strings.TrimSpace(req.Phone)

	if req.Name == "" || req.Address == "" || req.Email == "" || req.Phone == "" {
		return writeError(c, fiber.StatusBadRequest, "missing_fields", "name, address, email, and phone are required")
	}

	vendor, err := h.DB.UpdateVendor(context.Background(), db.UpdateVendorParams{
		ID:      id,
		Name:    req.Name,
		Address: req.Address,
		Email:   req.Email,
		Phone:   req.Phone,
	})
	if err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(vendor)
}

func (h *Handler) DeleteVendor(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	if err := h.DB.DeleteVendor(context.Background(), id); err != nil {
		return handleDBError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
