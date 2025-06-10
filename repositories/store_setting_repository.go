package repositories

import (
	"context"
	"go-fiber-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type StoreSettingRepository struct {
	collection *mongo.Collection
}

func NewStoreSettingRepository(db *mongo.Database) *StoreSettingRepository {
	return &StoreSettingRepository{
		collection: db.Collection("store_settings"),
	}
}

// Lấy setting đầu tiên
func (r *StoreSettingRepository) Get(ctx context.Context) (*models.StoreSetting, error) {
	var result models.StoreSetting
	err := r.collection.FindOne(ctx, bson.M{}).Decode(&result)
	return &result, err
}

// Ghi đè hoặc tạo mới (upsert)
func (r *StoreSettingRepository) Upsert(ctx context.Context, setting models.StoreSetting) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{}, bson.M{
		"$set": setting,
	}, mongo.UpdateOptions{
		Upsert: boolPtr(true),
	})
	return err
}

func boolPtr(b bool) *bool {
	return &b
}
