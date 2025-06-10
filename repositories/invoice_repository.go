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

// Create tạo hóa đơn mới, lưu thời gian theo GMT+7
func (r *InvoiceRepository) Create(ctx context.Context, invoice models.Invoice) error {
	loc, _ := time.LoadLocation("Asia/Bangkok") // GMT+7
	invoice.CreatedAt = time.Now().In(loc)
	_, err := r.collection.InsertOne(ctx, invoice)
	return err
}

// DeleteMany xoá nhiều hóa đơn theo ID
func (r *InvoiceRepository) DeleteMany(ctx context.Context, ids []string) error {
	var objIDs []primitive.ObjectID
	for _, id := range ids {
		objID, _ := primitive.ObjectIDFromHex(id)
		objIDs = append(objIDs, objID)
	}
	_, err := r.collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": objIDs}})
	return err
}

// ListByDateRange lọc hóa đơn theo khoảng ngày
func (r *InvoiceRepository) ListByDateRange(ctx context.Context, from, to time.Time) ([]models.Invoice, error) {
	filter := bson.M{
		"createdAt": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var result []models.Invoice
	if err = cursor.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListByDateRangePaginated lọc hóa đơn theo khoảng ngày + phân trang
func (r *InvoiceRepository) ListByDateRangePaginated(ctx context.Context, from, to time.Time, page, limit int64) ([]models.Invoice, int64, error) {
	filter := bson.M{
		"createdAt": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}

	opts := options.Find().
		SetSkip((page - 1) * limit).
		SetLimit(limit).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}

	var invoices []models.Invoice
	if err := cursor.All(ctx, &invoices); err != nil {
		return nil, 0, err
	}

	total, _ := r.collection.CountDocuments(ctx, filter)
	return invoices, total, nil
}

// Update cập nhật hóa đơn (tên cửa hàng, SĐT, sản phẩm, ghi chú)
func (r *InvoiceRepository) Update(ctx context.Context, id string, invoice models.Invoice) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"items": invoice.Items,
			"note":  invoice.Note,
		},
	}
	_, err = r.collection.UpdateByID(ctx, objID, update)
	return err
}
