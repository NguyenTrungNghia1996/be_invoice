package repositories

import (
	"context"

	"go-fiber-api/config"
	"go-fiber-api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// Tìm user theo username
func FindUserByUsername(username string) (*models.User, error) {
	var user models.User
	filter := bson.M{
		"username":  username,
		"deletedAt": bson.M{"$exists": false},
	}
	err := config.DB.Collection("users").FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser tạo mới một user
func CreateUser(user *models.User) error {
	user.ID = primitive.NewObjectID()
	_, err := config.DB.Collection("users").InsertOne(context.TODO(), user)
	return err
}

// Kiểm tra Username already exists
func IsUsernameExists(username string) (bool, error) {
	filter := bson.M{
		"username":  username,
		"deletedAt": bson.M{"$exists": false},
	}
	count, err := config.DB.Collection("users").CountDocuments(context.TODO(), filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Lấy danh sách user theo role (nếu có)
func GetUsersByRole(role string) ([]models.User, error) {
	filter := bson.M{"deletedAt": bson.M{"$exists": false}}
	if role != "" {
		filter["role"] = role
	}

	// Projection: loại bỏ trường password
	projection := bson.M{
		"password": 0, // 0 = không lấy trường này
	}

	opts := options.Find().SetProjection(projection)

	cursor, err := config.DB.Collection("users").Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var users []models.User
	for cursor.Next(context.TODO()) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func UpdateUserPassword(id string, hashedPassword string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objID, "deletedAt": bson.M{"$exists": false}}
	update := bson.M{"$set": bson.M{"password": hashedPassword}}
	_, err = config.DB.Collection("users").UpdateOne(context.TODO(), filter, update)
	return err
}

// UpdateUser cập nhật thông tin cơ bản của user (username, role)
func UpdateUser(id string, user models.User) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objID, "deletedAt": bson.M{"$exists": false}}
	update := bson.M{"$set": bson.M{
		"username": user.Username,
		"role":     user.Role,
	}}
	_, err = config.DB.Collection("users").UpdateOne(context.TODO(), filter, update)
	return err
}

// DeleteUsers xoá nhiều user theo danh sách ID
func DeleteUsers(ids []string) error {
	var objIDs []primitive.ObjectID
	for _, id := range ids {
		if objID, err := primitive.ObjectIDFromHex(id); err == nil {
			objIDs = append(objIDs, objID)
		}
	}
	if len(objIDs) == 0 {
		return nil
	}
	filter := bson.M{"_id": bson.M{"$in": objIDs}}
	update := bson.M{"$set": bson.M{"deletedAt": time.Now()}}
	_, err := config.DB.Collection("users").UpdateMany(context.TODO(), filter, update)
	return err
}

// Lấy user theo ID
func FindUserByID(id string) (*models.User, error) {
	var user models.User

	// ✅ Chuyển string sang ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objID, "deletedAt": bson.M{"$exists": false}}
	err = config.DB.Collection("users").FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
