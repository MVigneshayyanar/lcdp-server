package api

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/Maestrominds/lcdp-restaurant/internal/sqlc"
)

type menuItemRequest struct {
	Name        string              `json:"name"`
	Price       float64             `json:"price"`
	Category    string              `json:"category"`
	IsAvailable bool                `json:"is_available"`
	Description string              `json:"description"`
	Ingredients []ingredientRequest `json:"ingredients"`
}

type ingredientRequest struct {
	InventoryItemID int64   `json:"inventoryId"`
	Quantity        float64 `json:"quantity"`
}

func (h *Handler) CreateMenuItem(c *fiber.Ctx) error {
	var req menuItemRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || req.Price < 0 {
		return writeError(c, fiber.StatusBadRequest, "invalid_fields", "name and valid price are required")
	}

	ctx := context.Background()
	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		return handleDBError(c, err)
	}
	defer tx.Rollback(ctx)
	qtx := h.DB.WithTx(tx)

	item, err := qtx.CreateMenuItem(ctx, db.CreateMenuItemParams{
		Name:        req.Name,
		Price:       req.Price,
		IsAvailable: true,
		Category:    req.Category,
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
	})
	if err != nil {
		return handleDBError(c, err)
	}

	for _, ing := range req.Ingredients {
		_, err = qtx.CreateIngredient(ctx, db.CreateIngredientParams{
			MenuItemID:      item.ID,
			InventoryItemID: ing.InventoryItemID,
			Quantity:        ing.Quantity,
		})
		if err != nil {
			return handleDBError(c, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) ListMenuItems(c *fiber.Ctx) error {
	ctx := context.Background()
	items, err := h.DB.ListMenuItems(ctx)
	if err != nil {
		return handleDBError(c, err)
	}

	type menuItemResponse struct {
		db.MenuItem
		Ingredients []db.ListIngredientsByMenuItemRow `json:"ingredients"`
	}

	result := make([]menuItemResponse, len(items))
	for i, item := range items {
		ings, _ := h.DB.ListIngredientsByMenuItem(ctx, item.ID)
		if ings == nil {
			ings = []db.ListIngredientsByMenuItemRow{}
		}
		result[i] = menuItemResponse{
			MenuItem:    item,
			Ingredients: ings,
		}
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func (h *Handler) GetMenuItem(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	ctx := context.Background()
	item, err := h.DB.GetMenuItem(ctx, id)
	if err != nil {
		return handleDBError(c, err)
	}

	ings, _ := h.DB.ListIngredientsByMenuItem(ctx, id)
	if ings == nil {
		ings = []db.ListIngredientsByMenuItemRow{}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id":           item.ID,
		"name":         item.Name,
		"price":        item.Price,
		"is_available": item.IsAvailable,
		"category":     item.Category,
		"description":  item.Description,
		"created_at":   item.CreatedAt,
		"updated_at":   item.UpdatedAt,
		"ingredients":  ings,
	})
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

	ctx := context.Background()
	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		return handleDBError(c, err)
	}
	defer tx.Rollback(ctx)
	qtx := h.DB.WithTx(tx)

	item, err := qtx.UpdateMenuItem(ctx, db.UpdateMenuItemParams{
		ID:          id,
		Name:        req.Name,
		Price:       req.Price,
		IsAvailable: req.IsAvailable,
		Category:    req.Category,
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
	})
	if err != nil {
		return handleDBError(c, err)
	}

	// Update ingredients: delete old ones and insert new ones
	if err := qtx.DeleteIngredientsByMenuItem(ctx, id); err != nil {
		return handleDBError(c, err)
	}

	for _, ing := range req.Ingredients {
		_, err = qtx.CreateIngredient(ctx, db.CreateIngredientParams{
			MenuItemID:      id,
			InventoryItemID: ing.InventoryItemID,
			Quantity:        ing.Quantity,
		})
		if err != nil {
			return handleDBError(c, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(item)
}

func (h *Handler) DeleteMenuItem(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	ctx := context.Background()
	// First delete all ingredients associated with this menu item
	if err := h.DB.DeleteIngredientsByMenuItem(ctx, id); err != nil {
		return handleDBError(c, err)
	}

	if err := h.DB.DeleteMenuItem(ctx, id); err != nil {
		return handleDBError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
