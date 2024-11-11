package mongodb

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *MongoClient

func InitModelBase(client *MongoClient) {
	log.Println("init mongodb model base")
	mongoClient = client
}

type ModelEntitier interface {
	GetCreatedAt() time.Time
	SetCreatedAt(time.Time)
}

type EntityBase struct {
	CreatedAt time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

func (e *EntityBase) GetCreatedAt() time.Time {
	return e.CreatedAt
}

func (e *EntityBase) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

type EntityCounter struct {
	ID  string `bson:"_id"`
	Seq int64  `bson:"seq"`
}

type ModelBase[T any, PT ModelEntitier] struct {
	CollectionName string
}

func NewModelBase[T any, PT ModelEntitier](collectionName string) *ModelBase[T, PT] {
	return &ModelBase[T, PT]{CollectionName: collectionName}
}

func (mu *ModelBase[T, PT]) GetNextID() (int64, error) {
	filter := bson.D{{Key: "_id", Value: mu.CollectionName + "_id"}}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	collectionCounter, err := mongoClient.GetCollection("counters")
	if err != nil {
		return 0, err
	}

	var result EntityCounter
	err = collectionCounter.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&result)
	if err != nil {
		return 0, err
	}
	return result.Seq, nil
}

func (mu *ModelBase[T, PT]) NewEntity() *T {
	return new(T)
}

// 根据ID查找文档
func (mu *ModelBase[T, PT]) FindByID(ctx context.Context, id string) (*T, error) {
	collection, err := mongoClient.GetCollection(mu.CollectionName)
	if err != nil {
		return nil, err
	}

	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.D{{Key: "_id", Value: ID}}
	filter = append(filter, bson.E{Key: "deleted_at", Value: bson.D{{Key: "$exists", Value: false}}})
	var doc T
	err = collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &doc, nil
}

// 根据条件查找单个文档
func (mu *ModelBase[T, PT]) FindOne(ctx context.Context, filter bson.D) (*T, error) {
	collection, err := mongoClient.GetCollection(mu.CollectionName)
	if err != nil {
		return nil, err
	}

	filter = append(filter, bson.E{Key: "deleted_at", Value: bson.D{{Key: "$exists", Value: false}}})
	var doc T
	err = collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &doc, nil
}

// 根据条件查找单个文档并更新, 如果不存在则插入
func (mu *ModelBase[T, PT]) FindOneAndUpdate(ctx context.Context, filter bson.D, update bson.D, upsert bool) (*T, error) {
	collection, err := mongoClient.GetCollection(mu.CollectionName)
	if err != nil {
		return nil, err
	}

	filter = append(filter, bson.E{Key: "deleted_at", Value: bson.D{{Key: "$exists", Value: false}}})
	update = append(update, bson.E{Key: "$set", Value: bson.D{{Key: "updated_at", Value: time.Now()}}})
	update = append(update, bson.E{Key: "$setOnInsert", Value: bson.D{{Key: "created_at", Value: time.Now()}}})

	opts := options.FindOneAndUpdate().SetUpsert(upsert).SetReturnDocument(options.After)

	var doc T
	err = collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc)
	if err != nil { // 这里不用判断是不是空，统一返回错误
		return nil, err
	}

	return &doc, nil
}

// 根据条件查找多个文档
func (mu *ModelBase[T, PT]) FindMany(ctx context.Context, filter bson.D, findOptions *options.FindOptions) ([]*T, error) {
	collection, err := mongoClient.GetCollection(mu.CollectionName)
	if err != nil {
		return nil, err
	}

	filter = append(filter, bson.E{Key: "deleted_at", Value: bson.D{{Key: "$exists", Value: false}}})
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	var results []*T
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// 根据条件获取总数量
func (mu *ModelBase[T, PT]) Count(ctx context.Context, filter bson.D) (int64, error) {
	collection, err := mongoClient.GetCollection(mu.CollectionName)
	if err != nil {
		return 0, err
	}

	filter = append(filter, bson.E{Key: "deleted_at", Value: bson.D{{Key: "$exists", Value: false}}})
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// 插入单个文档
func (mu *ModelBase[T, PT]) InsertOne(ctx context.Context, doc PT) (string, error) {

	collection, err := mongoClient.GetCollection(mu.CollectionName)
	if err != nil {
		return "", err
	}

	if doc.GetCreatedAt().IsZero() {
		doc.SetCreatedAt(time.Now())
	}

	result, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return "", err
	}
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("invalid inserted id")
	}
	return id.Hex(), nil
}

// 插入多个文档
func (mu *ModelBase[T, PT]) InsertMany(ctx context.Context, docs []PT) ([]string, error) {

	collection, err := mongoClient.GetCollection(mu.CollectionName)
	if err != nil {
		return nil, err
	}

	var documents []interface{}
	for _, doc := range docs {
		if doc.GetCreatedAt().IsZero() {
			doc.SetCreatedAt(time.Now())
		}
		documents = append(documents, doc)
	}

	result, err := collection.InsertMany(ctx, documents)
	if err != nil {
		return nil, err
	}

	resultIDs := make([]string, 0, len(result.InsertedIDs))
	for _, id := range result.InsertedIDs {
		resultIDs = append(resultIDs, id.(primitive.ObjectID).Hex())
	}

	return resultIDs, nil
}

// 根据条件更新单个文档
func (mu *ModelBase[T, PT]) UpdateOne(ctx context.Context, filter bson.D, update bson.D) (*mongo.UpdateResult, error) {
	collection, err := mongoClient.GetCollection(mu.CollectionName)
	if err != nil {
		return nil, err
	}

	filter = append(filter, bson.E{Key: "deleted_at", Value: bson.D{{Key: "$exists", Value: false}}})
	update = append(update, bson.E{Key: "$set", Value: bson.D{{Key: "updated_at", Value: time.Now()}}})

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 根据ID更新单个文档
func (mu *ModelBase[T, PT]) UpdateByID(ctx context.Context, objectID string, update bson.D) (*mongo.UpdateResult, error) {
	collection, err := mongoClient.GetCollection(mu.CollectionName)
	if err != nil {
		return nil, err
	}

	update = append(update, bson.E{Key: "$set", Value: bson.D{{Key: "updated_at", Value: time.Now()}}})

	id, err := primitive.ObjectIDFromHex(objectID)
	if err != nil {
		return nil, err
	}
	result, err := collection.UpdateByID(ctx, id, update)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 根据条件更新多个文档
func (mu *ModelBase[T, PT]) UpdateMany(ctx context.Context, filter bson.D, update bson.D) (*mongo.UpdateResult, error) {
	collection, err := mongoClient.GetCollection(mu.CollectionName)
	if err != nil {
		return nil, err
	}

	filter = append(filter, bson.E{Key: "deleted_at", Value: bson.D{{Key: "$exists", Value: false}}})
	update = append(update, bson.E{Key: "$set", Value: bson.D{{Key: "updated_at", Value: time.Now()}}})

	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 软删除单个文档
// func (mu *ModelBase[T]) SoftDeleteOne(ctx context.Context, filter bson.D) (int64, error) {
// 	collection, err := mongoClient.GetCollection(mu.CollectionName)
// 	if err != nil {
// 		return 0, err
// 	}

// 	if len(filter) == 0 {
// 		return 0, errors.New("filter is empty")
// 	}

// 	update := bson.D{{Key: "$set", Value: bson.D{{Key: "deleted_at", Value: time.Now()}}}}
// 	result, err := collection.UpdateOne(ctx, filter, update)
// 	if err != nil {
// 		return 0, err
// 	}

// 	return result.ModifiedCount, nil
// }

// 硬删除单个文档
func (mu *ModelBase[T, PT]) DeleteOne(ctx context.Context, filter bson.D) (int64, error) {
	collection, err := mongoClient.GetCollection(mu.CollectionName)
	if err != nil {
		return 0, err
	}

	if len(filter) == 0 {
		return 0, errors.New("filter is empty")
	}

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

// 根据id软删除单个文档
// func (mu *ModelBase[T]) SoftDeleteByID(ctx context.Context, id string) error {
// 	collection, err := mongoClient.GetCollection(mu.CollectionName)
// 	if err != nil {
// 		return err
// 	}

// 	objectID, err := primitive.ObjectIDFromHex(id)
// 	if err != nil {
// 		return err
// 	}

// 	filter := bson.D{{Key: "_id", Value: objectID}}
// 	update := bson.D{{Key: "$set", Value: bson.D{{Key: "deleted_at", Value: time.Now()}}}}
// 	_, err = collection.UpdateOne(ctx, filter, update)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// 根据id硬删除单个文档
func (mu *ModelBase[T, PT]) DeleteByID(ctx context.Context, id string) (*mongo.DeleteResult, error) {
	collection, err := mongoClient.GetCollection(mu.CollectionName)
	if err != nil {
		return nil, err
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.D{{Key: "_id", Value: objectID}}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return result, err
	}

	return result, nil
}

// 根据条件软删除多个文档
// func (mu *ModelBase[T]) SoftDeleteMany(ctx context.Context, filter bson.D) (int64, error) {
// 	collection, err := mongoClient.GetCollection(mu.CollectionName)
// 	if err != nil {
// 		return 0, err
// 	}

// 	if len(filter) == 0 {
// 		return 0, errors.New("filter is empty")
// 	}

// 	update := bson.D{{Key: "$set", Value: bson.D{{Key: "deleted_at", Value: time.Now()}}}}
// 	result, err := collection.UpdateMany(ctx, filter, update)
// 	if err != nil {
// 		return 0, err
// 	}

// 	return result.ModifiedCount, nil
// }

// 根据条件硬删除多个文档
func (mu *ModelBase[T, PT]) DeleteMany(ctx context.Context, filter bson.D) (int64, error) {
	collection, err := mongoClient.GetCollection(mu.CollectionName)
	if err != nil {
		return 0, err
	}

	// 不允许删除所有数据
	if len(filter) == 0 {
		return 0, errors.New("filter is empty")
	}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}
