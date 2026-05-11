package api

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Maestrominds/lcdp-restaurant/internal/config"
	db "github.com/Maestrominds/lcdp-restaurant/internal/sqlc"
)

type Handler struct {
	DB     *db.Queries
	Config config.Config
	Pool   *pgxpool.Pool
}

func NewHandler(cfg config.Config, queries *db.Queries, pool *pgxpool.Pool) *Handler {
	return &Handler{DB: queries, Config: cfg, Pool: pool}
}

func (h *Handler) ListAlerts(c *fiber.Ctx) error {
	ctx := context.Background()
	var alerts []fiber.Map

	// 1. Low Stock Alerts (Generated on the fly)
	rows, err := h.Pool.Query(ctx, `
		SELECT id, name, quantity, min_stock, unit 
		FROM inventory_items 
		WHERE quantity <= min_stock
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int64
			var name, unit string
			var qty, min float64
			if err := rows.Scan(&id, &name, &qty, &min, &unit); err == nil {
				severity := "warning"
				if qty <= min/2 {
					severity = "critical"
				}
				alerts = append(alerts, fiber.Map{
					"id":          fmt.Sprintf("stock-%d", id),
					"type":        "critical",
					"title":       "Low Stock: " + name,
					"description": fmt.Sprintf("%s is running low (%.2f %s left, min %.0f)", name, qty, unit, min),
					"severity":    severity,
					"read":        false,
					"timestamp":   time.Now(),
				})
			}
		}
	}

	// 2. Overdue Bill Alerts
	billRows, err := h.Pool.Query(ctx, `
		SELECT b.id, v.name, b.amount, b.due_date 
		FROM bills b 
		JOIN vendors v ON b.vendor_id = v.id 
		WHERE b.status = 'pending' AND b.due_date < CURRENT_DATE
	`)
	if err == nil {
		defer billRows.Close()
		for billRows.Next() {
			var id int64
			var vendor string
			var amount float64
			var dueDate time.Time
			if err := billRows.Scan(&id, &vendor, &amount, &dueDate); err == nil {
				alerts = append(alerts, fiber.Map{
					"id":          fmt.Sprintf("bill-%d", id),
					"type":        "overdue",
					"title":       "Payment Overdue",
					"description": fmt.Sprintf("Invoice from %s (€%.2f) was due on %s", vendor, amount, dueDate.Format("2006-01-02")),
					"severity":    "warning",
					"read":        false,
					"timestamp":   time.Now(),
				})
			}
		}
	}

	// 3. Persistent Alerts from Database (If any)
	// For now, we only have auto-generated ones.
	
	if alerts == nil {
		alerts = []fiber.Map{}
	}

	return c.JSON(alerts)
}

func (h *Handler) MarkAlertRead(c *fiber.Ctx) error {
	// For now, since alerts are generated on the fly, marking as read doesn't persist.
	// In a full implementation, we would store the "read" status in the DB.
	return c.SendStatus(fiber.StatusOK)
}

func (h *Handler) Register(app *fiber.App) {
	// Base routes
	app.Get("/login", h.GetLogin)
	app.Post("/login/verify", h.VerifyLogin)

	// V1 Group for App/Website compatibility
	v1 := app.Group("/v1")
	
	// Public V1 routes (if any)
	v1.Post("/auth/manager/login", h.ManagerLogin)
	v1.Post("/auth/owner/login", h.OwnerLogin)
	v1.Post("/auth/otp/send", h.SendOtpFlutter)
	v1.Post("/auth/otp/verify", h.VerifyOtpFlutter)

	protected := v1.Group("", h.AuthMiddleware())

	// Analytics
	protected.Get("/analytics/overview", h.GetAnalyticsOverview)
	protected.Get("/analytics/revenue", h.GetAnalyticsRevenue)
	protected.Get("/analytics/sales", h.GetAnalyticsSales)
	protected.Get("/analytics/inventory", h.GetAnalyticsInventory)

	// Tables
	protected.Get("/tables", h.ListDiningTables)
	protected.Post("/tables", h.CreateDiningTable)
	protected.Get("/tables/:id", h.GetDiningTable)
	protected.Patch("/tables/:id/status", h.PatchDiningTableStatus)
	protected.Delete("/tables/:id", h.DeleteDiningTable)

	// Menu
	protected.Get("/menu", h.ListMenuItems)
	protected.Post("/menu", h.CreateMenuItem)
	protected.Get("/menu/:id", h.GetMenuItem)
	protected.Patch("/menu/:id", h.UpdateMenuItem)
	protected.Delete("/menu/:id", h.DeleteMenuItem)

	// Orders
	protected.Get("/orders", h.ListOrders)
	protected.Post("/orders", h.CreateOrder)
	protected.Get("/orders/:id", h.GetOrder)
	protected.Patch("/orders/:id/served", h.MarkOrderServed)
	protected.Delete("/orders/:id", h.DeleteOrder)

	// Inventory
	protected.Get("/inventory", h.ListInventoryItems)
	protected.Post("/inventory", h.CreateInventoryItem)
	protected.Get("/inventory/:id", h.GetInventoryItem)
	protected.Patch("/inventory/:id", h.UpdateInventoryItem)
	protected.Delete("/inventory/:id", h.DeleteInventoryItem)

	// Payables
	protected.Get("/payables", h.ListBills)
	protected.Patch("/payables/:id/paid", h.MarkPayablePaid)

	// Alerts
	protected.Get("/alerts", h.ListAlerts)
	protected.Patch("/alerts/:id/read", h.MarkAlertRead)

	// Kitchen
	protected.Get("/kitchen/orders", h.ListKitchenOrders)
	protected.Patch("/kitchen/orders/:id/start", h.StartKitchenOrder)
	protected.Patch("/kitchen/orders/:id/ready", h.ReadyKitchenOrder)

	// Waiters
	protected.Get("/waiters", h.ListUsers)
	protected.Post("/waiters", h.CreateUser)
	protected.Delete("/waiters/:id", h.DeleteUser)

	// Users
	protected.Get("/users", h.ListUsers)
	protected.Post("/users", h.CreateUser)

	// Vendors (full CRUD)
	protected.Get("/vendors", h.ListVendors)
	protected.Post("/vendors", h.CreateVendor)
	protected.Get("/vendors/:id", h.GetVendor)
	protected.Patch("/vendors/:id", h.UpdateVendor)
	protected.Delete("/vendors/:id", h.DeleteVendor)

	// Bills (full CRUD)
	protected.Get("/bills", h.ListBills)
	protected.Post("/bills", h.CreateBill)
	protected.Post("/bills/scan", h.ScanBill)
	protected.Get("/bills/:id", h.GetBill)
	protected.Patch("/bills/:id", h.UpdateBill)
	protected.Delete("/bills/:id", h.DeleteBill)

	// Reports
	protected.Get("/reports", h.ListReports)
}
