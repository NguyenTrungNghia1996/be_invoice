package repositories

import (
	"context"
	"fmt"
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

// generateInvoiceCode tạo mã hóa đơn dạng HD<YYYYMMDD><SEQ>
func generateInvoiceCode(db *mongo.Database) (string, error) {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc)
	dayKey := now.Format("20060102") // YYYYMMDD, ví dụ: 20250610
	counterID := fmt.Sprintf("invoice-%s", dayKey)

	filter := bson.M{"_id": counterID}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result struct {
		Seq int `bson:"seq"`
	}
	err := db.Collection("counters").FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&result)
	if err != nil {
		return "", err
	}

	code := fmt.Sprintf("HD%s%04d", dayKey, result.Seq) // HD202506100001
	return code, nil
}

// Create tạo hóa đơn mới, lưu thời gian theo GMT+7 và sinh mã hóa đơn tự động
func (r *InvoiceRepository) Create(ctx context.Context, invoice models.Invoice) error {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	invoice.CreatedAt = time.Now().In(loc)

	code, err := generateInvoiceCode(r.collection.Database())
	if err != nil {
		return err
	}
	invoice.Code = code

	_, err = r.collection.InsertOne(ctx, invoice)
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

// ListPaginated phân trang danh sách hóa đơn
func (r *InvoiceRepository) ListPaginated(ctx context.Context, page, limit int64) ([]models.Invoice, int64, error) {
	opts := options.Find().
		SetSkip((page - 1) * limit).
		SetLimit(limit).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}

	var invoices []models.Invoice
	if err := cursor.All(ctx, &invoices); err != nil {
		return nil, 0, err
	}

	total, _ := r.collection.CountDocuments(ctx, bson.M{})
	return invoices, total, nil
}

// Update cập nhật hóa đơn (sản phẩm, ghi chú)
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

// ListByCode lọc hóa đơn theo mã code (tìm gần đúng)
func (r *InvoiceRepository) ListByCode(ctx context.Context, code string) ([]models.Invoice, error) {
	filter := bson.M{
		"code": bson.M{
			"$regex": primitive.Regex{Pattern: code, Options: "i"},
		},
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var result []models.Invoice
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListByCodeAndDatePaginated lọc theo mã + ngày + phân trang
func (r *InvoiceRepository) ListByCodeAndDatePaginated(ctx context.Context, code string, from, to time.Time, page, limit int64) ([]models.Invoice, int64, error) {
	filter := bson.M{
		"code": bson.M{"$regex": primitive.Regex{Pattern: code, Options: "i"}},
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

