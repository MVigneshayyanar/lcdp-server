package api

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"

	db "github.com/Maestrominds/lcdp-restaurant/internal/sqlc"
)

type menuItemIngredientRequest struct {
	InventoryID int64   `json:"inventoryId"`
	Quantity    float64 `json:"quantity"`
}

type menuItemRequest struct {
	Name        string                       `json:"name"`
	Price       float64                      `json:"price"`
	Description *string                      `json:"description"`
	Category    string                       `json:"category"`
	SoldOut     bool                         `json:"soldOut"`
	IsAvailable *bool                        `json:"is_available"`
	Ingredients []menuItemIngredientRequest `json:"ingredients"`
}

func (h *Handler) CreateMenuItem(c *fiber.Ctx) error {
	var req menuItemRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return writeError(c, fiber.StatusBadRequest, "missing_fields", "Dish name is required")
	}
	if req.Price < 0 {
		return writeError(c, fiber.StatusBadRequest, "invalid_price", "Price cannot be negative")
	}
	if req.Category == "" {
		req.Category = "mains"
	}

	isAvailable := !req.SoldOut
	if req.IsAvailable != nil {
		isAvailable = *req.IsAvailable
	}

	ctx := context.Background()
	item, err := h.DB.CreateMenuItem(ctx, db.CreateMenuItemParams{
		Name:        req.Name,
		Price:       req.Price,
		IsAvailable: isAvailable,
		Category:    req.Category,
		Description: req.Description,
	})
	if err != nil {
		return handleDBError(c, err)
	}

	// Add ingredients
	for _, ing := range req.Ingredients {
		h.DB.CreateIngredient(ctx, db.CreateIngredientParams{
			MenuItemID:      item.ID,
			InventoryItemID: ing.InventoryID,
			Quantity:        ing.Quantity,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) ListMenuItems(c *fiber.Ctx) error {
	items, err := h.DB.ListMenuItems(context.Background())
	if err != nil {
		return handleDBError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(items)
}

func (h *Handler) GetMenuItem(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	item, err := h.DB.GetMenuItem(context.Background(), id)
	if err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(item)
}

func (h *Handler) UpdateMenuItem(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	var req menuItemRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return writeError(c, fiber.StatusBadRequest, "missing_fields", "Dish name is required")
	}
	if req.Price < 0 {
		return writeError(c, fiber.StatusBadRequest, "invalid_price", "Price cannot be negative")
	}
	if req.Category == "" {
		req.Category = "mains"
	}

	isAvailable := !req.SoldOut
	if req.IsAvailable != nil {
		isAvailable = *req.IsAvailable
	}

	ctx := context.Background()
	item, err := h.DB.UpdateMenuItem(ctx, db.UpdateMenuItemParams{
		ID:          id,
		Name:        req.Name,
		Price:       req.Price,
		IsAvailable: isAvailable,
		Category:    req.Category,
		Description: req.Description,
	})
	if err != nil {
		return handleDBError(c, err)
	}

	// Sync ingredients: Delete existing and add new
	h.DB.DeleteIngredientsByMenuItem(ctx, id)
	for _, ing := range req.Ingredients {
		h.DB.CreateIngredient(ctx, db.CreateIngredientParams{
			MenuItemID:      id,
			InventoryItemID: ing.InventoryID,
			Quantity:        ing.Quantity,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) DeleteMenuItem(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	if err := h.DB.DeleteMenuItem(context.Background(), id); err != nil {
		return handleDBError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
