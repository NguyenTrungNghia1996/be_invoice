package backup

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// StartAutoBackup sao lưu dữ liệu lên MongoDB Atlas định kỳ
func StartAutoBackup(ctx context.Context, localDB *mongo.Database) {
	atlasURI := os.Getenv("ATLAS_MONGO_URL")
	atlasName := os.Getenv("ATLAS_DB_NAME")
	if atlasURI == "" || atlasName == "" {
		log.Println("⚠️ ATLAS_MONGO_URL hoặc ATLAS_DB_NAME chưa cấu hình, bỏ qua backup")
		return
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := backupOnce(ctx, localDB, atlasURI, atlasName); err != nil {
					log.Printf("❌ Lỗi backup dữ liệu: %v\n", err)
				} else {
					log.Println("✅ Đã backup dữ liệu lên Atlas thành công")
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func backupOnce(ctx context.Context, localDB *mongo.Database, atlasURI, atlasName string) error {
	cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(cctx, options.Client().ApplyURI(atlasURI))
	if err != nil {
		return err
	}
	defer client.Disconnect(cctx)

	if err := client.Ping(cctx, nil); err != nil {
		return err
	}

	remoteDB := client.Database(atlasName)

	collections, err := localDB.ListCollectionNames(cctx, bson.M{})
	if err != nil {
		return err
	}

	for _, col := range collections {
		if err := copyCollection(cctx, localDB.Collection(col), remoteDB.Collection(col)); err != nil {
			return err
		}
	}
	return nil
}

func copyCollection(ctx context.Context, local, remote *mongo.Collection) error {
	cursor, err := local.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return err
		}
		id, ok := doc["_id"]
		if !ok {
			continue
		}
		_, err = remote.ReplaceOne(ctx, bson.M{"_id": id}, doc, options.Replace().SetUpsert(true))
		if err != nil {
			return err
		}
	}

	return cursor.Err()
}
