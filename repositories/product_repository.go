package repositories

import (
	"context"
	"errors"
	"go-fiber-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProductRepository struct {
	collection *mongo.Collection
}

func NewProductRepository(db *mongo.Database) *ProductRepository {
	return &ProductRepository{
		collection: db.Collection("products"),
	}
}

func (r *ProductRepository) Create(ctx context.Context, product models.Product) error {
	product.ID = primitive.NewObjectID()
	_, err := r.collection.InsertOne(ctx, product)
	return err
}

func (r *ProductRepository) Update(ctx context.Context, id string, product models.Product) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": product}
	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil || res.MatchedCount == 0 {
		return errors.New("update failed")
	}
	return nil
}

func (r *ProductRepository) DeleteMany(ctx context.Context, ids []string) error {
	var objIDs []primitive.ObjectID
	for _, id := range ids {
		objID, err := primitive.ObjectIDFromHex(id)
		if err == nil {
			objIDs = append(objIDs, objID)
		}
	}
	filter := bson.M{"_id": bson.M{"$in": objIDs}}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

func (r *ProductRepository) List(ctx context.Context, page, limit int64, search string) ([]models.Product, int64, error) {
	filter := bson.M{}
	if search != "" {
		filter["name"] = bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}}
	}

	opts := options.Find()
	if limit > 0 {
		opts.SetSkip((page - 1) * limit)
		opts.SetLimit(limit)
	}
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	var products []models.Product
	if err = cursor.All(ctx, &products); err != nil {
		return nil, 0, err
	}

	count, _ := r.collection.CountDocuments(ctx, filter)
	return products, count, nil
}
