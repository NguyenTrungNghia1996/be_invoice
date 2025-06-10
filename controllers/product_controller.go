package controllers

import (
	"github.com/gofiber/fiber/v2"
	"go-fiber-api/models"
	"go-fiber-api/repositories"
	"strings"
)

// ProductController là controller xử lý các API liên quan đến sản phẩm
type ProductController struct {
	repo *repositories.ProductRepository
}

// NewProductController khởi tạo controller với repository tương ứng
func NewProductController(repo *repositories.ProductRepository) *ProductController {
	return &ProductController{repo: repo}
}

// Create tạo mới một sản phẩm
// Method: POST /api/products
// Body JSON: { "name": "Sản phẩm A", "price": 10000 }
func (ctrl *ProductController) Create(c *fiber.Ctx) error {
	var product models.Product
	if err := c.BodyParser(&product); err != nil {
		return c.Status(400).JSON(models.APIResponse{"error", "Invalid input", nil})
	}
	err := ctrl.repo.Create(c.Context(), product)
	if err != nil {
		return c.Status(500).JSON(models.APIResponse{"error", "Create failed", nil})
	}
	return c.JSON(models.APIResponse{"success", "Created successfully", nil})
}

// Update cập nhật thông tin sản phẩm (lấy ID từ body)
// Method: PUT /api/products
// Body JSON: { "id": "abc123", "name": "Tên mới", "price": 15000 }
func (ctrl *ProductController) Update(c *fiber.Ctx) error {
	var product models.Product
	if err := c.BodyParser(&product); err != nil {
		return c.Status(400).JSON(models.APIResponse{"error", "Invalid input", nil})
	}

	if product.ID.IsZero() {
		return c.Status(400).JSON(models.APIResponse{"error", "Missing product ID", nil})
	}

	id := product.ID.Hex()
	err := ctrl.repo.Update(c.Context(), id, product)
	if err != nil {
		return c.Status(500).JSON(models.APIResponse{"error", "Update failed", nil})
	}
	return c.JSON(models.APIResponse{"success", "Updated successfully", nil})
}

// Delete xoá một hoặc nhiều sản phẩm
// Method: DELETE /api/products?id=abc123,def456
func (ctrl *ProductController) Delete(c *fiber.Ctx) error {
	ids := strings.Split(c.Query("id"), ",")
	err := ctrl.repo.DeleteMany(c.Context(), ids)
	if err != nil {
		return c.Status(500).JSON(models.APIResponse{"error", "Delete failed", nil})
	}
	return c.JSON(models.APIResponse{"success", "Deleted successfully", nil})
}

// List trả về danh sách sản phẩm có phân trang & tìm kiếm
// Method: GET /api/products?page=1&limit=10&search=tên
func (ctrl *ProductController) List(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search", "")
	data, total, err := ctrl.repo.List(c.Context(), int64(page), int64(limit), search)
	if err != nil {
		return c.Status(500).JSON(models.APIResponse{"error", "List failed", nil})
	}
	return c.JSON(models.APIResponse{"success", "List fetched", fiber.Map{
		"products": data,
		"page":     page,
		"limit":    limit,
		"total":    total,
	}})
}
