package controllers

import (
	"github.com/gofiber/fiber/v2"
	"go-fiber-api/models"
	"go-fiber-api/repositories"
	"strings"
	"time"
)

type InvoiceController struct {
	repo *repositories.InvoiceRepository
}

func NewInvoiceController(repo *repositories.InvoiceRepository) *InvoiceController {
	return &InvoiceController{repo: repo}
}

// Create tạo hóa đơn mới
//
// @route  POST /api/invoices
// @body   {
//   "storeName": "Shop ABC",
//   "phone": "0912345678",
//   "items": [
//     { "productId": "xxx", "name": "Áo sơ mi", "quantity": 2, "price": 150000 }
//   ],
//   "note": "Khách mua online"
// }
func (ctrl *InvoiceController) Create(c *fiber.Ctx) error {
	var invoice models.Invoice
	if err := c.BodyParser(&invoice); err != nil {
		return c.Status(400).JSON(models.APIResponse{"error", "Invalid input", nil})
	}
	if err := ctrl.repo.Create(c.Context(), invoice); err != nil {
		return c.Status(500).JSON(models.APIResponse{"error", "Create failed", nil})
	}
	return c.JSON(models.APIResponse{"success", "Invoice created", nil})
}

// Delete xoá một hoặc nhiều hóa đơn theo ID
//
// @route  DELETE /api/invoices?id=66a1...,66a2...
func (ctrl *InvoiceController) Delete(c *fiber.Ctx) error {
	ids := strings.Split(c.Query("id"), ",")
	if err := ctrl.repo.DeleteMany(c.Context(), ids); err != nil {
		return c.Status(500).JSON(models.APIResponse{"error", "Delete failed", nil})
	}
	return c.JSON(models.APIResponse{"success", "Invoices deleted", nil})
}

// FilterByDate lọc hóa đơn theo khoảng ngày (tùy chọn), mã code (tùy chọn), phân trang + thống kê
//
// @route  GET /api/invoices/filter?from=01/05/2025&to=31/05/2025&page=1&limit=10&code=HD20250610
func (ctrl *InvoiceController) FilterByDate(c *fiber.Ctx) error {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	code := c.Query("code")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	var fromTime, toTime time.Time
	filterByDate := fromStr != "" && toStr != ""
	filterByCode := code != ""

	if filterByDate {
		var err1, err2 error
		fromTime, err1 = time.ParseInLocation("02/01/2006", fromStr, time.FixedZone("GMT+7", 7*3600))
		toTime, err2 = time.ParseInLocation("02/01/2006", toStr, time.FixedZone("GMT+7", 7*3600))
		if err1 != nil || err2 != nil {
			return c.Status(400).JSON(models.APIResponse{"error", "Invalid date format (dd/mm/yyyy)", nil})
		}
		toTime = toTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}

	var (
		invoices []models.Invoice
		total    int64
		err      error
	)

	switch {
	case filterByDate && filterByCode:
		invoices, total, err = ctrl.repo.ListByCodeAndDatePaginated(c.Context(), code, fromTime, toTime, int64(page), int64(limit))
	case filterByDate:
		invoices, total, err = ctrl.repo.ListByDateRangePaginated(c.Context(), fromTime, toTime, int64(page), int64(limit))
	case filterByCode:
		invoices, err = ctrl.repo.ListByCode(c.Context(), code)
		total = int64(len(invoices))
	default:
		invoices, total, err = ctrl.repo.ListPaginated(c.Context(), int64(page), int64(limit))
	}

	if err != nil {
		return c.Status(500).JSON(models.APIResponse{"error", "List failed", nil})
	}

	// Thống kê sản phẩm trên kết quả trả về
	type ProductStats struct {
		Name     string  `json:"name"`
		Quantity int     `json:"quantity"`
		Revenue  float64 `json:"revenue"`
	}
	products := make(map[string]*ProductStats)
	var totalAmount float64

	for _, inv := range invoices {
		for _, item := range inv.Items {
			if _, ok := products[item.Name]; !ok {
				products[item.Name] = &ProductStats{Name: item.Name}
			}
			stat := products[item.Name]
			stat.Quantity += item.Quantity
			stat.Revenue += float64(item.Quantity) * item.Price
			totalAmount += float64(item.Quantity) * item.Price
		}
	}

	return c.JSON(models.APIResponse{"success", "Filtered invoices", fiber.Map{
		"invoices":     invoices,
		"page":         page,
		"limit":        limit,
		"total":        total,
		"totalAmount":  totalAmount,
		"productStats": products,
	}})
}

// Update cập nhật thông tin hóa đơn (sản phẩm, số lượng, cửa hàng, ghi chú)
//
// @route PUT /api/invoices
// @body  {
//   "id": "66ab...",
//   "storeName": "Shop XYZ",
//   "phone": "0909123456",
//   "items": [
//     { "productId": "xxx", "name": "Áo thun", "quantity": 1, "price": 120000 }
//   ],
//   "note": "Cập nhật đơn hàng"
// }
func (ctrl *InvoiceController) Update(c *fiber.Ctx) error {
	var invoice models.Invoice
	if err := c.BodyParser(&invoice); err != nil {
		return c.Status(400).JSON(models.APIResponse{
			Status:  "error",
			Message: "Invalid input",
			Data:    nil,
		})
	}
	if invoice.ID.IsZero() {
		return c.Status(400).JSON(models.APIResponse{
			Status:  "error",
			Message: "Missing invoice ID",
			Data:    nil,
		})
	}

	id := invoice.ID.Hex()
	if err := ctrl.repo.Update(c.Context(), id, invoice); err != nil {
		return c.Status(500).JSON(models.APIResponse{
			Status:  "error",
			Message: "Update failed",
			Data:    nil,
		})
	}

	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Invoice updated",
		Data:    nil,
	})
}
