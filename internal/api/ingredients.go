package api

import (
	"context"

	"github.com/gofiber/fiber/v2"

	db "github.com/Maestrominds/lcdp-restaurant/internal/sqlc"
)

type ingredientRequest struct {
	MenuItemID      int64   `json:"menu_item_id"`
	InventoryItemID int64   `json:"inventory_item_id"`
	Quantity        float64 `json:"quantity"`
}

type ingredientView struct {
	ID            int64   `json:"id"`
	MenuItem      string  `json:"menu_item"`
	InventoryItem string  `json:"inventory_item"`
	Quantity      float64 `json:"quantity"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func (h *Handler) CreateIngredient(c *fiber.Ctx) error {
	var req ingredientRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	if req.MenuItemID <= 0 || req.InventoryItemID <= 0 || req.Quantity <= 0 {
		return writeError(c, fiber.StatusBadRequest, "missing_fields", "menu_item_id, inventory_item_id, and quantity are required")
	}

	ingredient, err := h.DB.CreateIngredient(context.Background(), db.CreateIngredientParams{
		MenuItemID:      req.MenuItemID,
		InventoryItemID: req.InventoryItemID,
		Quantity:        req.Quantity,
	})
	if err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(ingredient)
}

func (h *Handler) ListIngredients(c *fiber.Ctx) error {
	ctx := context.Background()
	ingredients, err := h.DB.ListIngredients(ctx)
	if err != nil {
		return handleDBError(c, err)
	}

	menuItems, err := h.DB.ListMenuItems(ctx)
	if err != nil {
		return handleDBError(c, err)
	}
	menuNames := make(map[int64]string, len(menuItems))
	for _, item := range menuItems {
		menuNames[item.ID] = item.Name
	}

	invItems, err := h.DB.ListInventoryItems(ctx)
	if err != nil {
		return handleDBError(c, err)
	}
	invNames := make(map[int64]string, len(invItems))
	for _, item := range invItems {
		invNames[item.ID] = item.Name
	}

	views := make([]ingredientView, 0, len(ingredients))
	for _, ing := range ingredients {
		views = append(views, ingredientView{
			ID:            ing.ID,
			MenuItem:      menuNames[ing.MenuItemID],
			InventoryItem: invNames[ing.InventoryItemID],
			Quantity:      ing.Quantity,
			CreatedAt:     formatTime(ing.CreatedAt),
			UpdatedAt:     formatTime(ing.UpdatedAt),
		})
	}

	return c.Status(fiber.StatusOK).JSON(views)
}

func (h *Handler) GetIngredient(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	ctx := context.Background()
	ingredient, err := h.DB.GetIngredient(ctx, id)
	if err != nil {
		return handleDBError(c, err)
	}

	menuItem, err := h.DB.GetMenuItem(ctx, ingredient.MenuItemID)
	if err != nil {
		return handleDBError(c, err)
	}

	invItem, err := h.DB.GetInventoryItem(ctx, ingredient.InventoryItemID)
	if err != nil {
		return handleDBError(c, err)
	}

	view := ingredientView{
		ID:            ingredient.ID,
		MenuItem:      menuItem.Name,
		InventoryItem: invItem.Name,
		Quantity:      ingredient.Quantity,
		CreatedAt:     formatTime(ingredient.CreatedAt),
		UpdatedAt:     formatTime(ingredient.UpdatedAt),
	}

	return c.Status(fiber.StatusOK).JSON(view)
}

func (h *Handler) UpdateIngredient(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	var req ingredientRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	if req.MenuItemID <= 0 || req.InventoryItemID <= 0 || req.Quantity <= 0 {
		return writeError(c, fiber.StatusBadRequest, "missing_fields", "menu_item_id, inventory_item_id, and quantity are required")
	}

	ingredient, err := h.DB.UpdateIngredient(context.Background(), db.UpdateIngredientParams{
		ID:              id,
		MenuItemID:      req.MenuItemID,
		InventoryItemID: req.InventoryItemID,
		Quantity:        req.Quantity,
	})
	if err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(ingredient)
}

func (h *Handler) DeleteIngredient(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	if err := h.DB.DeleteIngredient(context.Background(), id); err != nil {
		return handleDBError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
