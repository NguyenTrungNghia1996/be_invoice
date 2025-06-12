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
	app.Post("/login", controllers.Login) // POST /login -> đăng nhập
	app.Get("/test", controllers.Hello)   // GET /test -> test không cần token
	api := app.Group("/api", middleware.Protected())
	api.Get("/test2", controllers.Hello)                // GET /api/test2 -> test có token
	api.Put("/presigned_url", controllers.GetUploadUrl) // PUT /api/presigned_url -> lấy URL upload ảnh (logo,...)

	usersGroup := api.Group("/users")

	usersGroup.Post("/", controllers.CreateUser)                // Tạo user mới
	usersGroup.Get("/", controllers.GetUsersByRole)             // Lấy danh sách user theo role (?role=)
	usersGroup.Put("/person", controllers.UpdateUserPersonID)   // Cập nhật person_id cho user
	usersGroup.Put("/password", controllers.ChangeUserPassword) // Đổi mật khẩu (kiểm tra mật khẩu cũ)
	// === Product routes ===
	productController := controllers.NewProductController(repositories.NewProductRepository(db))
	products := api.Group("/products")
	products.Get("/", productController.List)                              // GET /api/products?page=1&limit=10&search=abc -> danh sách sản phẩm
	products.Post("/", productController.Create)                           // POST /api/products -> tạo sản phẩm
	products.Put("/", productController.Update)                            // PUT /api/products -> cập nhật sản phẩm (ID trong body)
	products.Delete("/", middleware.AdminOnly(), productController.Delete) // DELETE /api/products?id=abc,def -> xóa nhiều sản phẩm

	// === Invoice routes ===
	invoiceController := controllers.NewInvoiceController(repositories.NewInvoiceRepository(db))
	invoices := api.Group("/invoices")
	invoices.Post("/", invoiceController.Create)                           // POST /api/invoices -> tạo hóa đơn
	invoices.Delete("/", middleware.AdminOnly(), invoiceController.Delete) // DELETE /api/invoices?id=abc,def -> xóa hóa đơn
	invoices.Get("/", invoiceController.FilterByDate)                      // GET /api/invoices?from=dd/mm/yyyy&to=dd/mm/yyyy&page=1&limit=10 -> lọc hóa đơn theo ngày
	invoices.Put("/", invoiceController.Update)                            // PUT /api/invoices -> cập nhật hóa đơn (ID trong body)

	// === Store setting routes ===
	settingCtrl := controllers.NewStoreSettingController(repositories.NewStoreSettingRepository(db))
	settings := api.Group("/settings")
	settings.Get("/", settingCtrl.Get)    // GET /api/settings -> lấy thông tin cửa hàng
	settings.Put("/", settingCtrl.Upsert) // PUT /api/settings -> cập nhật thông tin cửa hàng (tên, SĐT, logo)
}
