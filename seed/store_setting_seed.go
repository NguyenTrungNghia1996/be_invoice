package seed

import (
	"context"
	"fmt"
	"go-fiber-api/config"
	"go-fiber-api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func SeedStoreSettings() {
	collection := config.DB.Collection("store_settings")
	var existing models.StoreSetting
	err := collection.FindOne(context.TODO(), bson.M{}).Decode(&existing)
	if err != mongo.ErrNoDocuments {
		fmt.Println("✅ Store settings already exist.")
		return
	}

	setting := models.StoreSetting{
		StoreName: "Cửa hàng Trung Nghĩa",
		Phone:     "0911222333",
		LogoUrl:   "https://cdn.example.com/logo.png",
	}

	_, err = collection.InsertOne(context.TODO(), setting)
	if err != nil {
		fmt.Println("❌ Failed to seed store setting:", err)
		return
	}
	fmt.Println("🚀 Store settings seeded successfully.")
}
