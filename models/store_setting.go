package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type StoreSetting struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	StoreName string             `json:"storeName" bson:"storeName"`
	Phone     string             `json:"phone" bson:"phone"`
	LogoUrl   string             `json:"logoUrl" bson:"logoUrl"` // URL ảnh logo
}
