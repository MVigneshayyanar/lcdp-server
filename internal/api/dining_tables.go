package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	db "github.com/Maestrominds/lcdp-restaurant/internal/sqlc"
)

type diningTableRequest struct {
	Number int32  `json:"number"`
	Status string `json:"status"`
	Name   string `json:"name"`
	Seats  int32  `json:"seats"`
}

type diningTableStatusRequest struct {
	Status string `json:"status"`
}

var allowedDiningTableStatuses = map[string]struct{}{
	"available": {},
	"ordered":   {},
	"preparing": {},
	"ready":      {},
	"eating":    {},
	"billed":    {},
}

func isValidDiningTableStatus(status string) bool {
	_, ok := allowedDiningTableStatuses[status]
	return ok
}

func (h *Handler) CreateDiningTable(c *fiber.Ctx) error {
	var req diningTableRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Status = strings.TrimSpace(req.Status)
	
	if req.Number <= 0 && req.Name != "" {
		// Try to extract number from name (e.g., "Table 5" -> 5)
		fmt.Sscanf(req.Name, "Table %d", &req.Number)
	}

	if req.Number <= 0 {
		req.Number = int32(time.Now().UnixNano() % 1000000)
	}

	if req.Status == "" {
		req.Status = "available"
	}

	if !isValidDiningTableStatus(req.Status) {
		return writeError(c, fiber.StatusBadRequest, "invalid_status", "Status must be available, ordered, preparing, or eating")
	}

	table, err := h.DB.CreateDiningTable(context.Background(), db.CreateDiningTableParams{
		Number: req.Number,
		Status: db.TableStatus(req.Status),
		Name:   req.Name,
		Seats:  req.Seats,
	})
	if err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(table)
}

func (h *Handler) ListDiningTables(c *fiber.Ctx) error {
	tables, err := h.DB.ListDiningTables(context.Background())
	if err != nil {
		return handleDBError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(tables)
}

func (h *Handler) GetDiningTable(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	table, err := h.DB.GetDiningTable(context.Background(), id)
	if err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(table)
}

func (h *Handler) DeleteDiningTable(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	if err := h.DB.DeleteDiningTable(context.Background(), id); err != nil {
		return handleDBError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) PatchDiningTableStatus(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	var req diningTableStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		return writeError(c, fiber.StatusBadRequest, "missing_fields", "status is required")
	}

	if !isValidDiningTableStatus(req.Status) {
		return writeError(c, fiber.StatusBadRequest, "invalid_status", "invalid dining table status")
	}

	ctx := context.Background()
	current, err := h.DB.GetDiningTable(ctx, id)
	if err != nil {
		return handleDBError(c, err)
	}

	if current.Status == db.TableStatus(status) {
		return c.Status(fiber.StatusOK).JSON(current)
	}

	updated, err := h.DB.UpdateDiningTableStatus(ctx, id, db.TableStatus(status))
	if err != nil {
		return handleDBError(c, err)
	}

	if updated.Status == db.TableStatus("preparing") {
		order, err := h.DB.ListOrders(ctx)
		if err != nil {
			return handleDBError(c, err)
		}

		var currentOrder db.Order
		for _, o := range order {
			if o.TableID == id {
				currentOrder = o
				break
			}
		}

		if currentOrder == (db.Order{}) {
			return writeError(c, fiber.StatusConflict, "no_order", "no order found for this table")
		}

		ingredients, err := h.DB.ListIngredients(ctx)
		if err != nil {
			return handleDBError(c, err)
		}

		menuIngredients := make(map[int64][]db.Ingredient)
		for _, ingredient := range ingredients {
			menuIngredients[ingredient.MenuItemID] = append(menuIngredients[ingredient.MenuItemID], ingredient)
		}

		inventoryItems, err := h.DB.ListInventoryItems(ctx)
		if err != nil {
			return handleDBError(c, err)
		}

		invByID := make(map[int64]db.InventoryItem, len(inventoryItems))
		qtyByID := make(map[int64]float64, len(inventoryItems))
		for _, item := range inventoryItems {
			invByID[item.ID] = item
			qtyByID[item.ID] = item.Quantity
		}

		changed := make(map[int64]struct{})
		for _, ingredient := range menuIngredients[currentOrder.MenuItemID] {
			currentQty, ok := qtyByID[ingredient.InventoryItemID]
			if !ok {
				return writeError(c, fiber.StatusConflict, "invalid_inventory", "inventory item not found")
			}
			needed := ingredient.Quantity * float64(currentOrder.Quantity)
			newQty := currentQty - needed
			if newQty < 0 {
				inv := invByID[ingredient.InventoryItemID]
				return writeError(c, fiber.StatusConflict, "insufficient_inventory", "insufficient inventory for "+inv.Name)
			}
			qtyByID[ingredient.InventoryItemID] = newQty
			changed[ingredient.InventoryItemID] = struct{}{}
		}

		for id := range changed {
			item := invByID[id]
			_, err := h.DB.UpdateInventoryItem(ctx, db.UpdateInventoryItemParams{
				ID:       item.ID,
				Name:     item.Name,
				Quantity: qtyByID[id],
				Unit:     item.Unit,
				VendorID: item.VendorID,
			})
			if err != nil {
				return handleDBError(c, err)
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(updated)
}
