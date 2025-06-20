package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string             `bson:"username" json:"username"`
	Password  string             `bson:"password,omitempty" json:"password,omitempty"`
	Role      string             `bson:"role" json:"role"`
	DeletedAt *time.Time         `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`
}
