package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/Maestrominds/lcdp-restaurant/internal/sqlc"
)

type billRequest struct {
	VendorID int64   `json:"vendor_id"`
	TxnID    string  `json:"txn_id"`
	Amount   float64 `json:"amount"`
	DueDate  string  `json:"due_date"`
	Status   string  `json:"status"`
}

type billView struct {
	ID            int64   `json:"id"`
	Vendor        string  `json:"vendor"`
	BillReference string  `json:"billReference"`
	Amount        float64 `json:"amount"`
	DueDate       string  `json:"dueDate"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

var allowedBillStatuses = map[string]struct{}{
	"pending": {},
	"paid":    {},
	"overdue": {},
}

func isValidBillStatus(status string) bool {
	_, ok := allowedBillStatuses[status]
	return ok
}

func (h *Handler) CreateBill(c *fiber.Ctx) error {
	var req billRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	req.TxnID = strings.TrimSpace(req.TxnID)
	req.Status = strings.TrimSpace(req.Status)

	if req.VendorID <= 0 || req.TxnID == "" || req.Amount <= 0 || req.DueDate == "" || req.Status == "" {
		return writeError(c, fiber.StatusBadRequest, "missing_fields", "vendor_id, txn_id, amount, due_date, and status are required")
	}

	if !isValidBillStatus(req.Status) {
		return writeError(c, fiber.StatusBadRequest, "invalid_status", "invalid bill status")
	}

	dueDate, err := parseDate(req.DueDate)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_date", "due_date must be YYYY-MM-DD")
	}

	bill, err := h.DB.CreateBill(context.Background(), db.CreateBillParams{
		VendorID: req.VendorID,
		TxnID:    req.TxnID,
		Amount:   req.Amount,
		DueDate:  pgtype.Date{Time: dueDate, Valid: true},
		Status:   db.BillStatus(req.Status),
	})
	if err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(bill)
}

func (h *Handler) ListBills(c *fiber.Ctx) error {
	ctx := context.Background()
	bills, err := h.DB.ListBills(ctx)
	if err != nil {
		return handleDBError(c, err)
	}

	vendors, err := h.DB.ListVendors(ctx)
	if err != nil {
		return handleDBError(c, err)
	}

	vendorNames := make(map[int64]string, len(vendors))
	for _, vendor := range vendors {
		vendorNames[vendor.ID] = vendor.Name
	}

	views := make([]billView, 0, len(bills))
	for _, bill := range bills {
		views = append(views, billView{
			ID:            bill.ID,
			Vendor:        vendorNames[bill.VendorID],
			BillReference: bill.TxnID,
			Amount:        bill.Amount,
			DueDate:       formatDate(bill.DueDate),
			Status:        string(bill.Status),
			CreatedAt:     formatTime(bill.CreatedAt),
			UpdatedAt:     formatTime(bill.UpdatedAt),
		})
	}

	return c.Status(fiber.StatusOK).JSON(views)
}

func (h *Handler) GetBill(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	ctx := context.Background()
	bill, err := h.DB.GetBill(ctx, id)
	if err != nil {
		return handleDBError(c, err)
	}

	vendor, err := h.DB.GetVendor(ctx, bill.VendorID)
	if err != nil {
		return handleDBError(c, err)
	}

	view := billView{
		ID:            bill.ID,
		Vendor:        vendor.Name,
		BillReference: bill.TxnID,
		Amount:        bill.Amount,
		DueDate:       formatDate(bill.DueDate),
		Status:        string(bill.Status),
		CreatedAt:     formatTime(bill.CreatedAt),
		UpdatedAt:     formatTime(bill.UpdatedAt),
	}

	return c.Status(fiber.StatusOK).JSON(view)
}

func (h *Handler) UpdateBill(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	var req billRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	req.TxnID = strings.TrimSpace(req.TxnID)
	req.Status = strings.TrimSpace(req.Status)

	if req.VendorID <= 0 || req.TxnID == "" || req.Amount <= 0 || req.DueDate == "" || req.Status == "" {
		return writeError(c, fiber.StatusBadRequest, "missing_fields", "vendor_id, txn_id, amount, due_date, and status are required")
	}

	if !isValidBillStatus(req.Status) {
		return writeError(c, fiber.StatusBadRequest, "invalid_status", "invalid bill status")
	}

	dueDate, err := parseDate(req.DueDate)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_date", "due_date must be YYYY-MM-DD")
	}

	bill, err := h.DB.UpdateBill(context.Background(), db.UpdateBillParams{
		ID:       id,
		VendorID: req.VendorID,
		TxnID:    req.TxnID,
		Amount:   req.Amount,
		DueDate:  pgtype.Date{Time: dueDate, Valid: true},
		Status:   db.BillStatus(req.Status),
	})
	if err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(bill)
}

func (h *Handler) DeleteBill(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	if err := h.DB.DeleteBill(context.Background(), id); err != nil {
		return handleDBError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) MarkPayablePaid(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	_, err = h.Pool.Exec(context.Background(), "UPDATE bills SET status = 'paid', updated_at = now() WHERE id = $1", id)
	if err != nil {
		return handleDBError(c, err)
	}

	return c.SendStatus(fiber.StatusOK)
}
func (h *Handler) ScanBill(c *fiber.Ctx) error {
	type scanRequest struct {
		Image    string `json:"image"`
		MimeType string `json:"mimeType"`
	}
	var req scanRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	// In a real app, this would call an OCR service or AI model.
	// For this demo, we'll return a mock scanned bill result.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"vendor_id": 1,
		"txn_id":    fmt.Sprintf("SCAN-%d", time.Now().Unix()),
		"amount":    124.50,
		"due_date":   time.Now().AddDate(0, 0, 30).Format("2006-01-02"),
		"status":     "pending",
	})
}
