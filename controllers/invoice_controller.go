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
//   "items": [
//     { "productId": "66a1...", "name": "Áo sơ mi", "quantity": 2, "price": 150000 },
//     { "productId": "66a2...", "name": "Quần jeans", "quantity": 1, "price": 300000 }
//   ],
//   "note": "Khách mua vào sáng thứ 2"
// }
//
// ✅ Dữ liệu trả về:
// {
//   "status": "success",
//   "message": "Invoice created",
//   "data": null
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

// Delete xóa một hoặc nhiều hóa đơn theo ID
//
// @route  DELETE /api/invoices?id=66a1...,66a2...
//
// ✅ Dữ liệu trả về:
// {
//   "status": "success",
//   "message": "Invoices deleted",
//   "data": null
// }
func (ctrl *InvoiceController) Delete(c *fiber.Ctx) error {
	ids := strings.Split(c.Query("id"), ",")
	if err := ctrl.repo.DeleteMany(c.Context(), ids); err != nil {
		return c.Status(500).JSON(models.APIResponse{"error", "Delete failed", nil})
	}
	return c.JSON(models.APIResponse{"success", "Invoices deleted", nil})
}

// FilterByDate lọc hóa đơn theo khoảng ngày + phân trang + thống kê
//
// @route  GET /api/invoices/filter?from=01/05/2025&to=31/05/2025&page=1&limit=10
//
// ✅ Dữ liệu trả về:
// {
//   "status": "success",
//   "message": "Filtered invoices",
//   "data": {
//     "invoices": [
//       {
//         "id": "66ab...",
//         "createdAt": "2025-05-10T10:20:00+07:00",
//         "items": [
//           { "productId": "66a1...", "name": "Áo sơ mi", "quantity": 2, "price": 150000 }
//         ],
//         "note": "Khách A"
//       }
//     ],
//     "page": 1,
//     "limit": 10,
//     "total": 13,
//     "totalAmount": 300000,
//     "productStats": {
//       "Áo sơ mi": { "name": "Áo sơ mi", "quantity": 5, "revenue": 750000 },
//       "Quần jeans": { "name": "Quần jeans", "quantity": 3, "revenue": 900000 }
//     }
//   }
// }
func (ctrl *InvoiceController) FilterByDate(c *fiber.Ctx) error {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	fromTime, err1 := time.ParseInLocation("02/01/2006", fromStr, time.FixedZone("GMT+7", 7*3600))
	toTime, err2 := time.ParseInLocation("02/01/2006", toStr, time.FixedZone("GMT+7", 7*3600))
	if err1 != nil || err2 != nil {
		return c.Status(400).JSON(models.APIResponse{"error", "Invalid date format (dd/mm/yyyy)", nil})
	}
	toTime = toTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	invoices, total, err := ctrl.repo.ListByDateRangePaginated(c.Context(), fromTime, toTime, int64(page), int64(limit))
	if err != nil {
		return c.Status(500).JSON(models.APIResponse{"error", "List failed", nil})
	}

	// Thống kê theo sản phẩm
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
