package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	db "github.com/Maestrominds/lcdp-restaurant/internal/sqlc"
)

type inventoryItemRequest struct {
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
	Vendor   string  `json:"vendor"`
	VendorID int64   `json:"vendorId"`
	Category string  `json:"category"`
	MinStock float64 `json:"minStock"`
}

type inventoryItemView struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	Vendor      string  `json:"vendor"`
	VendorID    int64   `json:"vendorId"`
	Category    string  `json:"category"`
	MinStock    float64 `json:"minStock"`
	Status      string  `json:"status"` // critical, low, ok
	LastUpdated string  `json:"lastUpdated"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func (h *Handler) CreateInventoryItem(c *fiber.Ctx) error {
	var req inventoryItemRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Unit = strings.TrimSpace(req.Unit)
	
	if req.Name == "" || req.Unit == "" {
		return writeError(c, fiber.StatusBadRequest, "missing_fields", "Name and Unit are required")
	}

	if req.Quantity < 0 {
		return writeError(c, fiber.StatusBadRequest, "invalid_quantity", "Quantity cannot be negative")
	}

	ctx := context.Background()
	
	// Resolve Vendor
	req.Vendor = strings.TrimSpace(req.Vendor)
	if req.Vendor != "" {
		// Prioritize the name string to allow editing vendors via text input
		vendors, _ := h.DB.ListVendors(ctx)
		found := false
		for _, v := range vendors {
			if strings.EqualFold(v.Name, req.Vendor) {
				req.VendorID = v.ID
				found = true
				break
			}
		}
		if !found {
			v, err := h.DB.CreateVendor(ctx, db.CreateVendorParams{
				Name:    req.Vendor,
				Address: "N/A",
				Email:   fmt.Sprintf("vendor-%d@cafedeparis.com", time.Now().UnixNano()),
				Phone:   fmt.Sprintf("%d", time.Now().UnixNano())[:10],
				Status:  "active",
			})
			if err != nil {
				return writeError(c, fiber.StatusInternalServerError, "vendor_creation_failed", "Failed to create new vendor: "+err.Error())
			}
			req.VendorID = v.ID
		}
	} else if req.VendorID <= 0 {
		// Fallback only if both name and ID are missing
		vendors, _ := h.DB.ListVendors(ctx)
		if len(vendors) > 0 {
			req.VendorID = vendors[0].ID
		} else {
			v, err := h.DB.CreateVendor(ctx, db.CreateVendorParams{
				Name:    "General Vendor",
				Address: "N/A",
				Email:   fmt.Sprintf("general-%d@cafedeparis.com", time.Now().UnixNano()),
				Phone:   fmt.Sprintf("%d", time.Now().UnixNano())[:10],
				Status:  "active",
			})
			if err != nil {
				return writeError(c, fiber.StatusInternalServerError, "vendor_creation_failed", "Failed to create default vendor: "+err.Error())
			}
			req.VendorID = v.ID
		}
	}

	if req.Category == "" {
		req.Category = "Pantry"
	}

	item, err := h.DB.CreateInventoryItem(ctx, db.CreateInventoryItemParams{
		Name:     req.Name,
		Quantity: req.Quantity,
		Unit:     req.Unit,
		VendorID: req.VendorID,
		Category: req.Category,
		MinStock: req.MinStock,
	})
	if err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) ListInventoryItems(c *fiber.Ctx) error {
	ctx := context.Background()
	items, err := h.DB.ListInventoryItems(ctx)
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

	views := make([]inventoryItemView, 0, len(items))
	for _, item := range items {
		status := "ok"
		if item.Quantity <= item.MinStock/2 {
			status = "critical"
		} else if item.Quantity <= item.MinStock {
			status = "low"
		}

		views = append(views, inventoryItemView{
			ID:          item.ID,
			Name:        item.Name,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			Vendor:      vendorNames[item.VendorID],
			VendorID:    item.VendorID,
			Category:    item.Category,
			MinStock:    item.MinStock,
			Status:      status,
			LastUpdated: formatTime(item.UpdatedAt),
			CreatedAt:   formatTime(item.CreatedAt),
			UpdatedAt:   formatTime(item.UpdatedAt),
		})
	}

	return c.Status(fiber.StatusOK).JSON(views)
}

func (h *Handler) GetInventoryItem(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	ctx := context.Background()
	item, err := h.DB.GetInventoryItem(ctx, id)
	if err != nil {
		return handleDBError(c, err)
	}

	vendor, err := h.DB.GetVendor(ctx, item.VendorID)
	if err != nil {
		return handleDBError(c, err)
	}

	status := "ok"
	if item.Quantity <= item.MinStock/2 {
		status = "critical"
	} else if item.Quantity <= item.MinStock {
		status = "low"
	}

	view := inventoryItemView{
		ID:          item.ID,
		Name:        item.Name,
		Quantity:    item.Quantity,
		Unit:        item.Unit,
		Vendor:      vendor.Name,
		VendorID:    item.VendorID,
		Category:    item.Category,
		MinStock:    item.MinStock,
		Status:      status,
		LastUpdated: formatTime(item.UpdatedAt),
		CreatedAt:   formatTime(item.CreatedAt),
		UpdatedAt:   formatTime(item.UpdatedAt),
	}

	return c.Status(fiber.StatusOK).JSON(view)
}

func (h *Handler) UpdateInventoryItem(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	var req inventoryItemRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_body", "invalid json body")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Unit = strings.TrimSpace(req.Unit)
	if req.Name == "" || req.Unit == "" {
		return writeError(c, fiber.StatusBadRequest, "missing_fields", "Name and Unit are required")
	}

	ctx := context.Background()
	
	// Resolve Vendor
	req.Vendor = strings.TrimSpace(req.Vendor)
	if req.Vendor != "" {
		// Prioritize the name string to allow editing vendors via text input
		vendors, _ := h.DB.ListVendors(ctx)
		found := false
		for _, v := range vendors {
			if strings.EqualFold(v.Name, req.Vendor) {
				req.VendorID = v.ID
				found = true
				break
			}
		}
		if !found {
			v, err := h.DB.CreateVendor(ctx, db.CreateVendorParams{
				Name:    req.Vendor,
				Address: "N/A",
				Email:   fmt.Sprintf("vendor-%d@cafedeparis.com", time.Now().UnixNano()),
				Phone:   fmt.Sprintf("%d", time.Now().UnixNano())[:10],
				Status:  "active",
			})
			if err != nil {
				return writeError(c, fiber.StatusInternalServerError, "vendor_creation_failed", "Failed to create new vendor: "+err.Error())
			}
			req.VendorID = v.ID
		}
	} else if req.VendorID <= 0 {
		// Fallback only if both name and ID are missing
		vendors, _ := h.DB.ListVendors(ctx)
		if len(vendors) > 0 {
			req.VendorID = vendors[0].ID
		} else {
			v, err := h.DB.CreateVendor(ctx, db.CreateVendorParams{
				Name:    "General Vendor",
				Address: "N/A",
				Email:   fmt.Sprintf("general-%d@cafedeparis.com", time.Now().UnixNano()),
				Phone:   fmt.Sprintf("%d", time.Now().UnixNano())[:10],
				Status:  "active",
			})
			if err != nil {
				return writeError(c, fiber.StatusInternalServerError, "vendor_creation_failed", "Failed to create default vendor: "+err.Error())
			}
			req.VendorID = v.ID
		}
	}

	item, err := h.DB.UpdateInventoryItem(ctx, db.UpdateInventoryItemParams{
		ID:       id,
		Name:     req.Name,
		Quantity: req.Quantity,
		Unit:     req.Unit,
		VendorID: req.VendorID,
		Category: req.Category,
		MinStock: req.MinStock,
	})
	if err != nil {
		return handleDBError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(item)
}

func (h *Handler) DeleteInventoryItem(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid_id", err.Error())
	}

	if err := h.DB.DeleteInventoryItem(context.Background(), id); err != nil {
		return handleDBError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
