package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"` // ✅ đúng kiểu
	Username string             `bson:"username" json:"username"`
	Password string             `bson:"password,omitempty" json:"-"`
	Role     string             `bson:"role" json:"role"`
	PersonID string             `bson:"person_id" json:"personID"`
}
