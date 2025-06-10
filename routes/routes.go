package routes

import (
	"go-fiber-api/controllers"
	"go-fiber-api/middleware"
	"go-fiber-api/repositories"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// Setup cấu hình toàn bộ route cho ứng dụng
func Setup(app *fiber.App, db *mongo.Database) {
	// === Public routes ===
	// POST /login -> đăng nhập
	app.Post("/login", controllers.Login)

	// GET /test -> test không cần token
	app.Get("/test", controllers.Hello)

	// === Protected API routes ===
	api := app.Group("/api", middleware.Protected())

	// GET /api/test2 -> test có token
	api.Get("/test2", controllers.Hello)

	// PUT /api/presigned_url -> lấy URL upload ảnh (logo,...)
	api.Put("/presigned_url", controllers.GetUploadUrl)

	// === Product routes ===
	productController := controllers.NewProductController(repositories.NewProductRepository(db))
	products := api.Group("/products")

	// GET /api/products?page=1&limit=10&search=abc -> danh sách sản phẩm
	products.Get("/", productController.List)

	// POST /api/products -> tạo sản phẩm
	products.Post("/", productController.Create)

	// PUT /api/products -> cập nhật sản phẩm (ID trong body)
	products.Put("/", productController.Update)

	// DELETE /api/products?id=abc,def -> xóa nhiều sản phẩm
	products.Delete("/", productController.Delete)

	// === Invoice routes ===
	invoiceController := controllers.NewInvoiceController(repositories.NewInvoiceRepository(db))
	invoices := api.Group("/invoices")

	// POST /api/invoices -> tạo hóa đơn
	invoices.Post("/", invoiceController.Create)

	// DELETE /api/invoices?id=abc,def -> xóa hóa đơn
	invoices.Delete("/", invoiceController.Delete)

	// GET /api/invoices/filter?from=dd/mm/yyyy&to=dd/mm/yyyy&page=1&limit=10 -> lọc hóa đơn theo ngày
	invoices.Get("/filter", invoiceController.FilterByDate)

	// PUT /api/invoices -> cập nhật hóa đơn (ID trong body)
	invoices.Put("/", invoiceController.Update)

	// === Store setting routes ===
	settingCtrl := controllers.NewStoreSettingController(repositories.NewStoreSettingRepository(db))
	settings := api.Group("/settings")

	// GET /api/settings -> lấy thông tin cửa hàng
	settings.Get("/", settingCtrl.Get)

	// PUT /api/settings -> cập nhật thông tin cửa hàng (tên, SĐT, logo)
	settings.Put("/", settingCtrl.Upsert)}


