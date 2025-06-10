package repositories

import (
	"context"
	"time"

	"go-fiber-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type InvoiceRepository struct {
	collection *mongo.Collection
}

func NewInvoiceRepository(db *mongo.Database) *InvoiceRepository {
	return &InvoiceRepository{
		collection: db.Collection("invoices"),
	}
}

func (r *InvoiceRepository) Create(ctx context.Context, invoice models.Invoice) error {
	loc, _ := time.LoadLocation("Asia/Bangkok") // GMT+7
	invoice.CreatedAt = time.Now().In(loc)
	_, err := r.collection.InsertOne(ctx, invoice)
	return err
}

func (r *InvoiceRepository) DeleteMany(ctx context.Context, ids []string) error {
	var objIDs []primitive.ObjectID
	for _, id := range ids {
		objID, _ := primitive.ObjectIDFromHex(id)
		objIDs = append(objIDs, objID)
	}
	_, err := r.collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": objIDs}})
	return err
}

func (r *InvoiceRepository) ListByDateRangePaginated(ctx context.Context, from, to time.Time, page, limit int64) ([]models.Invoice, int64, error) {
	filter := bson.M{
		"createdAt": bson.M{
