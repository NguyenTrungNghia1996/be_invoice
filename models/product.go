package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Product struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	Price     float64            `json:"price" bson:"price"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
}
