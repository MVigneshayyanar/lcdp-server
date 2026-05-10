package api

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/Maestrominds/lcdp-restaurant/internal/sqlc"
)

type orderItemReq struct {
	MenuItemID interface{} `json:"menuItemId"`
	Quantity   int32       `json:"quantity"`
}

// Aliases for snake_case compatibility
type orderItemReqSnake struct {
	MenuItemID interface{} `json:"menu_item_id"`
	Quantity   int32       `json:"quantity"`
}

type orderRequest struct {
	TableID interface{}    `json:"tableId"`
	Items   []orderItemReq `json:"items"`
}

type orderRequestSnake struct {
	TableID interface{}         `json:"table_id"`
	Items   []orderItemReqSnake `json:"items"`
}

type orderViewItem struct {
	Name     string `json:"name"`
	Quantity int32  `json:"quantity"`
}

type orderView struct {
	ID        int64            `json:"id"`
	TableID   int64            `json:"tableId"`
	TableName string           `json:"tableName"`
	Status    string           `json:"status"`
	Items     []orderViewItem  `json:"items"`
	CreatedAt string           `json:"createdAt"`
}

func (h *Handler) CreateOrder(c *fiber.Ctx) error {
	fmt.Println("[DEBUG] CreateOrder hit!")
	fmt.Printf("[DEBUG] Raw Body: %s\n", string(c.Body()))

	var req orderRequest
	if err := c.BodyParser(&req); err != nil {
		// Try snake_case if camelCase fails
		var reqSnake orderRequestSnake
		if err2 := c.BodyParser(&reqSnake); err2 == nil {
			req.TableID = reqSnake.TableID
			req.Items = make([]orderItemReq, len(reqSnake.Items))
			for i, it := range reqSnake.Items {
				req.Items[i] = orderItemReq{MenuItemID: it.MenuItemID, Quantity: it.Quantity}
			}
		} else {
			return writeError(c, 400, "invalid_request", err.Error())
		}
	}
	fmt.Printf("[DEBUG] Request: %+v\n", req)

	tableID := toInt64(req.TableID)
	if tableID == 0 {
		return writeError(c, 400, "invalid_table_id", "Table ID is required")
	}

	if len(req.Items) == 0 {
		return writeError(c, 400, "empty_order", "Order must contain at least one item")
	}

	ctx := context.Background()
	// Use a transaction for atomic order creation
	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		return handleDBError(c, err)
	}
	defer tx.Rollback(ctx)

	qtx := h.DB.WithTx(tx)
	var createdOrders []db.Order

	for _, item := range req.Items {
		menuItemID := toInt64(item.MenuItemID)
		if menuItemID == 0 {
			continue
		}

		order, err := qtx.CreateOrder(ctx, db.CreateOrderParams{
			MenuItemID: menuItemID,
			Quantity:   item.Quantity,
			TableID:    tableID,
			OrderedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
			Status:     db.OrderStatusNew,
		})
		if err != nil {
			fmt.Printf("[ERROR] CreateOrder failed for table %d, item %d: %v\n", tableID, menuItemID, err)
			return handleDBError(c, err)
		}
		createdOrders = append(createdOrders, order)
	}

	if len(createdOrders) > 0 {
		// Update table status to 'ordered'
		_, err = qtx.UpdateDiningTableStatus(ctx, tableID, db.TableStatusOrdered)
		if err != nil {
			fmt.Printf("[ERROR] UpdateDiningTableStatus failed for table %d: %v\n", tableID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "Success", "count": len(createdOrders)})
}

func (h *Handler) ListOrders(c *fiber.Ctx) error {
	ctx := context.Background()
	orders, err := h.DB.ListOrders(ctx)
	if err != nil {
		return handleDBError(c, err)
	}

	menuItems, _ := h.DB.ListMenuItems(ctx)
	menuNames := make(map[int64]string)
	for _, m := range menuItems {
		menuNames[m.ID] = m.Name
	}

	tables, _ := h.DB.ListDiningTables(ctx)
	tableNames := make(map[int64]string)
	for _, t := range tables {
		tableNames[t.ID] = t.Name
	}

	// Group by TableID and Status
	type groupKey struct {
		TableID int64
		Status  db.OrderStatus
	}
	groups := make(map[groupKey]*orderView)
	var keys []groupKey

	for _, o := range orders {
		key := groupKey{o.TableID, o.Status}
		if _, ok := groups[key]; !ok {
			keys = append(keys, key)
			groups[key] = &orderView{
				ID:        o.ID,
				TableID:   o.TableID,
				TableName: tableNames[o.TableID],
				Status:    string(o.Status),
				CreatedAt: formatTime(o.CreatedAt),
				Items:     []orderViewItem{},
			}
		}
		groups[key].Items = append(groups[key].Items, orderViewItem{
			Name:     menuNames[o.MenuItemID],
			Quantity: o.Quantity,
		})
	}

	result := make([]*orderView, 0, len(keys))
	for _, k := range keys {
		result = append(result, groups[k])
	}

	return c.JSON(result)
}

func (h *Handler) GetOrder(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	ctx := context.Background()
	order, err := h.DB.GetOrder(ctx, id)
	if err != nil {
		return handleDBError(c, err)
	}

	menuItem, err := h.DB.GetMenuItem(ctx, order.MenuItemID)
	if err != nil {
		return handleDBError(c, err)
	}

	table, err := h.DB.GetDiningTable(ctx, order.TableID)
	if err != nil {
		return handleDBError(c, err)
	}

	view := orderView{
		ID:        order.ID,
		TableID:   order.TableID,
		TableName: table.Name,
		Status:    string(order.Status),
		CreatedAt: formatTime(order.CreatedAt),
		Items: []orderViewItem{
			{Name: menuItem.Name, Quantity: order.Quantity},
		},
	}

	return c.Status(fiber.StatusOK).JSON(view)
}

func (h *Handler) DeleteOrder(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}
	if err := h.DB.DeleteOrder(context.Background(), id); err != nil {
		return handleDBError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
func (h *Handler) ListKitchenOrders(c *fiber.Ctx) error {
	ctx := context.Background()
	orders, err := h.DB.ListOrders(ctx)
	if err != nil {
		return handleDBError(c, err)
	}

	menuItems, _ := h.DB.ListMenuItems(ctx)
	menuNames := make(map[int64]string)
	for _, m := range menuItems {
		menuNames[m.ID] = m.Name
	}

	tables, _ := h.DB.ListDiningTables(ctx)
	tableNames := make(map[int64]string)
	for _, t := range tables {
		tableNames[t.ID] = t.Name
	}

	type kitchenOrder struct {
		ID        int64    `json:"id"`
		Table     string   `json:"table"`
		Status    string   `json:"status"`
		CreatedAt int64    `json:"createdAt"`
		Items     []string `json:"items"`
	}

	// Group by Table and Status
	type kKey struct {
		TableID int64
		Status  db.OrderStatus
	}
	groups := make(map[kKey]*kitchenOrder)
	var keys []kKey

	for _, o := range orders {
		if o.Status == db.OrderStatusServed {
			continue
		}
		key := kKey{o.TableID, o.Status}
		if _, ok := groups[key]; !ok {
			keys = append(keys, key)
			groups[key] = &kitchenOrder{
				ID:        o.ID,
				Table:     tableNames[o.TableID],
				Status:    string(o.Status),
				CreatedAt: o.CreatedAt.Time.Unix() * 1000,
				Items:     []string{},
			}
		}
		groups[key].Items = append(groups[key].Items, fmt.Sprintf("%dx %s", o.Quantity, menuNames[o.MenuItemID]))
	}

	result := make([]*kitchenOrder, 0, len(keys))
	for _, k := range keys {
		result = append(result, groups[k])
	}

	return c.JSON(result)
}

func (h *Handler) MarkOrderServed(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, 400, "invalid_id", err.Error())
	}
	ctx := context.Background()
	order, err := h.DB.GetOrder(ctx, id)
	if err != nil {
		return handleDBError(c, err)
	}

	// Update ALL orders for this table with the same status (ready -> served)
	allOrders, _ := h.DB.ListOrders(ctx)
	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		return handleDBError(c, err)
	}
	defer tx.Rollback(ctx)
	qtx := h.DB.WithTx(tx)

	for _, o := range allOrders {
		if o.TableID == order.TableID && o.Status == order.Status {
			_, err = qtx.UpdateOrder(ctx, db.UpdateOrderParams{
				ID:         o.ID,
				MenuItemID: o.MenuItemID,
				Quantity:   o.Quantity,
				TableID:    o.TableID,
				OrderedAt:  o.OrderedAt,
				Status:     db.OrderStatusServed,
			})
			if err != nil {
				fmt.Printf("[ERROR] MarkOrderServed failed for order %d: %v\n", o.ID, err)
			}
		}
	}

	// Update table status to 'eating'
	qtx.UpdateDiningTableStatus(ctx, order.TableID, db.TableStatusEating)

	if err := tx.Commit(ctx); err != nil {
		return handleDBError(c, err)
	}

	return c.Status(200).JSON(fiber.Map{"status": "success"})
}

func (h *Handler) StartKitchenOrder(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, 400, "invalid_id", err.Error())
	}
	ctx := context.Background()
	order, err := h.DB.GetOrder(ctx, id)
	if err != nil {
		return handleDBError(c, err)
	}

	// Update ALL orders for this table with the same status (new -> preparing)
	allOrders, _ := h.DB.ListOrders(ctx)
	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		return handleDBError(c, err)
	}
	defer tx.Rollback(ctx)
	qtx := h.DB.WithTx(tx)

	for _, o := range allOrders {
		if o.TableID == order.TableID && o.Status == order.Status {
			_, err = qtx.UpdateOrder(ctx, db.UpdateOrderParams{
				ID:         o.ID,
				MenuItemID: o.MenuItemID,
				Quantity:   o.Quantity,
				TableID:    o.TableID,
				OrderedAt:  o.OrderedAt,
				Status:     db.OrderStatusPreparing,
			})
			if err != nil {
				fmt.Printf("[ERROR] StartKitchenOrder failed for order %d: %v\n", o.ID, err)
			}
		}
	}

	// Update table status to 'preparing'
	qtx.UpdateDiningTableStatus(ctx, order.TableID, db.TableStatusPreparing)

	// Subtract inventory for all items being prepared
	ingredients, _ := h.DB.ListIngredients(ctx)
	invItems, _ := h.DB.ListInventoryItems(ctx)
	invByID := make(map[int64]db.InventoryItem)
	for _, it := range invItems {
		invByID[it.ID] = it
	}

	// Track cumulative deductions
	deductions := make(map[int64]float64)
	for _, o := range allOrders {
		if o.TableID == order.TableID && o.Status == order.Status {
			for _, ing := range ingredients {
				if ing.MenuItemID == o.MenuItemID {
					deductions[ing.InventoryItemID] += ing.Quantity * float64(o.Quantity)
				}
			}
		}
	}

	for invID, needed := range deductions {
		inv := invByID[invID]
		qtx.UpdateInventoryItem(ctx, db.UpdateInventoryItemParams{
			ID:       inv.ID,
			Name:     inv.Name,
			Quantity: inv.Quantity - needed,
			Unit:     inv.Unit,
			VendorID: inv.VendorID,
			Category: inv.Category,
			MinStock: inv.MinStock,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return handleDBError(c, err)
	}

	return c.Status(200).JSON(fiber.Map{"status": "success"})
}

func (h *Handler) ReadyKitchenOrder(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, 400, "invalid_id", err.Error())
	}
	ctx := context.Background()
	order, err := h.DB.GetOrder(ctx, id)
	if err != nil {
		return handleDBError(c, err)
	}

	// Update ALL orders for this table with the same status (preparing -> ready)
	allOrders, _ := h.DB.ListOrders(ctx)
	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		return handleDBError(c, err)
	}
	defer tx.Rollback(ctx)
	qtx := h.DB.WithTx(tx)

	for _, o := range allOrders {
		if o.TableID == order.TableID && o.Status == order.Status {
			_, err = qtx.UpdateOrder(ctx, db.UpdateOrderParams{
				ID:         o.ID,
				MenuItemID: o.MenuItemID,
				Quantity:   o.Quantity,
				TableID:    o.TableID,
				OrderedAt:  o.OrderedAt,
				Status:     db.OrderStatusReady,
			})
			if err != nil {
				fmt.Printf("[ERROR] ReadyKitchenOrder failed for order %d: %v\n", o.ID, err)
			}
		}
	}

	// Update table status to 'ready'
	qtx.UpdateDiningTableStatus(ctx, order.TableID, db.TableStatusReady)

	if err := tx.Commit(ctx); err != nil {
		return handleDBError(c, err)
	}

	return c.Status(200).JSON(fiber.Map{"status": "success"})
}


func (h *Handler) ListReports(c *fiber.Ctx) error {
	return c.JSON([]fiber.Map{
		{"id": "rep1", "name": "Daily Sales Report", "description": "Complete breakdown of today's sales and revenue", "frequency": "Daily", "generatedAt": time.Now(), "size": "1.2 MB"},
		{"id": "rep2", "name": "Inventory Usage Report", "description": "Weekly tracking of stock consumption", "frequency": "Weekly", "generatedAt": time.Now().AddDate(0, 0, -1), "size": "2.5 MB"},
	})
}
