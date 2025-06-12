package seed

import (
	"context"
	"fmt"
	"go-fiber-api/config"
	"go-fiber-api/models"
	"go-fiber-api/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func seedUserIfNotExists(username, password, role string) {
	collection := config.DB.Collection("users")

	// Kiểm tra user đã tồn tại chưa
	var existing models.User
	err := collection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&existing)
	if err != mongo.ErrNoDocuments {
		fmt.Printf("✅ User '%s' already exists.\n", username)
		return
	}

	// Băm mật khẩu
	hashedPwd, err := utils.HashPassword(password)
	if err != nil {
		fmt.Printf("❌ Failed to hash password for '%s': %v\n", username, err)
		return
	}

	user := models.User{
		Username: username,
		Password: hashedPwd,
		Role:     role,
	}

	// Tạo mới người dùng
	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		fmt.Printf("❌ Failed to seed user '%s': %v\n", username, err)
		return
	}

	fmt.Printf("🚀 Seeded user successfully: username=%s, password=%s, role=%s\n", username, password, role)
}

func SeedAdminUser() {
	seedUserIfNotExists("admin", "admin123", "admin")
	seedUserIfNotExists("user", "user123", "user")
}
