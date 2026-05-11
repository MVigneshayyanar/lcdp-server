package api

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"

	db "github.com/Maestrominds/lcdp-restaurant/internal/sqlc"
)

type diningTableRequest struct {
	Number int32 `json:"number"`
}

func (h *Handler) CreateDiningTable(c *fiber.Ctx) error {
	var req diningTableRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	if req.Number <= 0 {
		req.Number = int32(time.Now().UnixNano() % 1000000)
	}

	table, err := h.DB.CreateDiningTable(context.Background(), req.Number)
	if err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(table)
}

func (h *Handler) ListDiningTables(c *fiber.Ctx) error {
	ctx := context.Background()
	tables, err := h.DB.ListDiningTables(ctx)
	if err != nil {
		return handleDBError(c, err)
	}

	// Dynamic status calculation
	orders, _ := h.DB.ListOrders(ctx)
	tableStatus := make(map[int64]string)
	
	// Create a map of tables for quick access to UpdatedAt
	tableMap := make(map[int64]db.DiningTable)
	for _, t := range tables {
		tableMap[t.ID] = t
	}

	for _, o := range orders {
		t, ok := tableMap[o.TableID]
		if !ok { continue }
		
		// If the order was placed BEFORE the table was last "cleared" or "freed", ignore it
		if o.OrderedAt.Time.Before(t.UpdatedAt.Time) {
			continue
		}

		if o.Status != db.OrderStatusServed {
			tableStatus[o.TableID] = string(o.Status)
		} else if tableStatus[o.TableID] == "" {
			tableStatus[o.TableID] = "eating"
		}
	}

	type tableResponse struct {
		ID        int64  `json:"id"`
		Number    int32  `json:"number"`
		Name      string `json:"name"`
		Status    string `json:"status"`
		CreatedAt string `json:"createdAt"`
	}

	res := make([]tableResponse, len(tables))
	for i, t := range tables {
		status := t.Status
		if status == "" || status == "available" || status == "ordered" {
			// Fallback to order status if available/ordered but has active orders
			orderStatus := tableStatus[t.ID]
			if orderStatus != "" {
				status = orderStatus
			}
			// Removed the logic that forced "ordered" back to "available" 
			// if no active orders were found, as the DB status is authoritative.
		}
		
		res[i] = tableResponse{
			ID:        t.ID,
			Number:    t.Number,
			Name:      fmt.Sprintf("Table %d", t.Number),
			Status:    status,
			CreatedAt: t.CreatedAt.Time.Format(time.RFC3339),
		}
	}

	return c.Status(fiber.StatusOK).JSON(res)
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

	var req struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	_, err = h.Pool.Exec(context.Background(), "UPDATE dining_tables SET status = $1, updated_at = now() WHERE id = $2", req.Status, id)
	if err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "table status updated"})
}
