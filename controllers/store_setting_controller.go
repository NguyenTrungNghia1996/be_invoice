package controllers

import (
	"github.com/gofiber/fiber/v2"
	"go-fiber-api/models"
	"go-fiber-api/repositories"
)

type StoreSettingController struct {
	repo *repositories.StoreSettingRepository
}

func NewStoreSettingController(repo *repositories.StoreSettingRepository) *StoreSettingController {
	return &StoreSettingController{repo: repo}
}

// GET /api/settings
// ✅ Trả về:
// {
//   "status": "success",
//   "message": "Fetched store info",
//   "data": {
//     "storeName": "Cửa hàng A",
//     "phone": "0912345678",
//     "logoUrl": "https://..."
//   }
// }
func (ctrl *StoreSettingController) Get(c *fiber.Ctx) error {
	setting, err := ctrl.repo.Get(c.Context())
	if err != nil {
		return c.Status(500).JSON(models.APIResponse{"error", "Get failed", nil})
	}
	return c.JSON(models.APIResponse{"success", "Fetched store info", setting})
}

// PUT /api/settings
// Body:
// {
//   "storeName": "Cửa hàng mới",
//   "phone": "0909123456",
//   "logoUrl": "https://cdn.com/logo.png"
// }
func (ctrl *StoreSettingController) Upsert(c *fiber.Ctx) error {
	var setting models.StoreSetting
	if err := c.BodyParser(&setting); err != nil {
		return c.Status(400).JSON(models.APIResponse{"error", "Invalid input", nil})
	}
	if err := ctrl.repo.Upsert(c.Context(), setting); err != nil {
		return c.Status(500).JSON(models.APIResponse{"error", "Update failed", nil})
	}
	return c.JSON(models.APIResponse{"success", "Store info saved", nil})
}
