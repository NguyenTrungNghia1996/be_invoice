package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Invoice struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"` // Giờ GMT+7
	Items     []InvoiceItem      `json:"items" bson:"items"`
	Note      string             `json:"note" bson:"note,omitempty"`
}

type InvoiceItem struct {
	ProductID primitive.ObjectID `json:"productId" bson:"productId"`
	Name      string             `json:"name" bson:"name"`
	Quantity  int                `json:"quantity" bson:"quantity"`
	Price     float64            `json:"price" bson:"price"` // đơn giá
}
