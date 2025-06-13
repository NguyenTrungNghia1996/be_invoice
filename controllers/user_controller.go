package controllers

import (
	"go-fiber-api/models"
	"go-fiber-api/repositories"
	"go-fiber-api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"os"
	"strings"
)

// CreateUser handles the creation of a new user
// POST /api/users
// Body:
//
//	{
//	  "username": "teacher1",
//	  "password": "123456",
//	  "email": "teacher@example.com",
//	  "role": "member",
//	}
func CreateUser(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Invalid data",
			Data:    nil,
		})
	}

	// Kiểm tra username đã tồn tại chưa
	exists, err := repositories.IsUsernameExists(user.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Error checking username",
			Data:    nil,
		})
	}
	if exists {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Username already exists",
			Data:    nil,
		})
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Failed to hash password",
			Data:    nil,
		})
	}
	user.Password = hashedPassword

	// Tạo user
	if err := repositories.CreateUser(&user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Unable to create user",
			Data:    nil,
		})
	}

	// Xoá password trước khi trả về frontend
	user.Password = ""

	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Created user successfully",
		Data:    user,
	})
}

// GetUsersByRole retrieves users by their role
// GET /api/users?role=member
func GetUsersByRole(c *fiber.Ctx) error {
	role := c.Query("role")

	users, err := repositories.GetUsersByRole(role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Cannot get user list",
			Data:    nil,
		})
	}

	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Get user list successfully",
		Data:    users,
	})
}

// ChangeUserPassword updates a user's password after verifying the old one
// PUT /api/users/password
// Body:
//
//	{
//	  "id": "665e1b3fa6ef0c2d7e3e594f",
//	  "old_password": "admin123",
//	  "new_password": "123456"
//	}
func ChangeUserPassword(c *fiber.Ctx) error {
	var body struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	// Gọi BodyParser đúng 1 lần
	err := c.BodyParser(&body)
	if err != nil || body.OldPassword == "" || body.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Status:  "error",
			Message: "Invalid data",
			Data:    nil,
		})
	}

	// Lấy token từ header
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIResponse{
			Status:  "error",
			Message: "Missing or invalid Authorization header",
			Data:    nil,
		})
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Kiểm tra signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "Unexpected signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIResponse{
			Status:  "error",
			Message: "Invalid token",
			Data:    nil,
		})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIResponse{
			Status:  "error",
			Message: "Invalid token claims",
			Data:    nil,
		})
	}

	userID, ok := claims["id"].(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIResponse{
			Status:  "error",
			Message: "User ID not found in token",
			Data:    nil,
		})
	}
	// Tìm user
	user, err := repositories.FindUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.APIResponse{
			Status:  "error",
			Message: "User not found",
			Data:    nil,
		})
	}

	// Kiểm tra mật khẩu cũ
	if !utils.CheckPasswordHash(body.OldPassword, user.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIResponse{
			Status:  "error",
			Message: "Old password is incorrect",
			Data:    nil,
		})
	}

	// Mã hoá mật khẩu mới
	hashed, err := utils.HashPassword(body.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Unable to encrypt password",
			Data:    nil,
		})
	}

	// Cập nhật mật khẩu
	err = repositories.UpdateUserPassword(userID, hashed)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Status:  "error",
			Message: "Unable to update password",
			Data:    nil,
		})
	}

	return c.JSON(models.APIResponse{
		Status:  "success",
		Message: "Password changed successfully",
		Data:    nil,
	})
}

// UpdateUser cập nhật thông tin cơ bản của người dùng
//
// @route PUT /api/users
// @body
//
//	{
//	    "id": "665e1b3fa6ef0c2d7e3e594f",
//	    "username": "newname",
//	    "email": "new@example.com",
//	    "role": "member"
//	}
func UpdateUser(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil || user.ID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{Status: "error", Message: "Invalid data", Data: nil})
	}
	if err := repositories.UpdateUser(user.ID, user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{Status: "error", Message: "Unable to update user", Data: nil})
	}
	return c.JSON(models.APIResponse{Status: "success", Message: "User updated", Data: nil})
}

// DeleteUsers xoá một hoặc nhiều người dùng
//
// @route DELETE /api/users?id=abc,def
func DeleteUsers(c *fiber.Ctx) error {
	ids := strings.Split(c.Query("id"), ",")
	if len(ids) == 0 || ids[0] == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{Status: "error", Message: "Missing id", Data: nil})
	}
	if err := repositories.DeleteUsers(ids); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{Status: "error", Message: "Delete failed", Data: nil})
	}
	return c.JSON(models.APIResponse{Status: "success", Message: "Deleted successfully", Data: nil})
}
