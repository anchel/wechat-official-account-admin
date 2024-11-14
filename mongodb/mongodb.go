package mongodb

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	client   *mongo.Client
	database *mongo.Database
	lock     sync.RWMutex
}

var (
	instance *MongoClient
	once     sync.Once
)

type MongoConfig struct {
	Username string
	Password string
	Host     string
	Database string
}

func NewMongoConfig() *MongoConfig {
	return &MongoConfig{
		Username: os.Getenv("MONGO_USER"),
		Password: os.Getenv("MONGO_PASSWORD"),
		Host:     os.Getenv("MONGO_HOST"),
		Database: os.Getenv("MONGO_DB"),
	}
}

func NewMongoClientOptions() *options.ClientOptions {
	connectTimeout := 6 * time.Second
	var maxPoolSize uint64 = 10
	cfg := NewMongoConfig()
	return options.Client().
		SetAuth(options.Credential{
			Username:   cfg.Username,
			Password:   cfg.Password,
			AuthSource: cfg.Database,
		}).
		SetHosts([]string{cfg.Host}).
		SetConnectTimeout(connectTimeout).
		SetMaxPoolSize(maxPoolSize)
}

// NewMongoClient 初始化并返回 MongoClient 的单例实例
func NewMongoClient() (*MongoClient, error) {
	var err error

	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cfg := NewMongoConfig()
		clientOpts := NewMongoClientOptions()

		client, connErr := mongo.Connect(ctx, clientOpts)
		if connErr != nil {
			err = connErr
			return
		}

		if pingErr := client.Ping(ctx, nil); pingErr != nil {
			err = pingErr
			return
		}

		instance = &MongoClient{
			client:   client,
			database: client.Database(cfg.Database),
		}

		log.Println("MongoDB connection established.")
	})

	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (m *MongoClient) GetDatabase(dbName string) *mongo.Database {
	return m.client.Database(dbName)
}

// GetCollection 获取指定集合的句柄
func (m *MongoClient) GetCollection(collectionName string) (*mongo.Collection, error) {
	if m.database == nil {
		return nil, errors.New("database not initialized")
	}
	return m.database.Collection(collectionName), nil
}

// Reconnect 重连 MongoDB，适用于连接失败的情况
func (m *MongoClient) Reconnect(ctx context.Context) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if err := m.client.Disconnect(ctx); err != nil {
		log.Printf("Failed to disconnect: %v", err)
		return err
	}

	clientOpts := NewMongoClientOptions()

	newClient, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Printf("Failed to reconnect to MongoDB: %v", err)
		return err
	}

	m.client = newClient
	m.database = newClient.Database(m.database.Name())
	log.Println("Reconnected to MongoDB successfully.")
	return nil
}

// Disconnect 优雅地断开连接
func (m *MongoClient) Disconnect(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// HealthCheck 检查 MongoDB 的健康状态
func (m *MongoClient) HealthCheck(ctx context.Context) bool {
	if err := m.client.Ping(ctx, nil); err != nil {
		log.Printf("MongoDB is not healthy: %v", err)
		return false
	}
	return true
}

func GetCollectionIndexs(ctx context.Context, collection *mongo.Collection) ([]bson.M, error) {
	indexView := collection.Indexes()
	opts := options.ListIndexes().SetMaxTime(8 * time.Second)
	cursor, err := indexView.List(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Get a slice of all indexes returned and print them out.
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// CheckCollectionIndexExists 检查集合中是否存在指定的索引，单字段索引
func CheckCollectionIndexExists(indexs []bson.M, fieldName string, checkUnique bool) bool {
	for _, index := range indexs {
		if keys, ok := index["key"].(bson.M); ok {
			if len(keys) == 1 {
				if _, exists := keys[fieldName]; exists {
					if checkUnique {
						if unique, ok := index["unique"].(bool); ok && unique {
							return true
						}
					} else {
						return true
					}
				}
			}
		}
	}
	return false
}

// CheckCollectionCompoundIndexExists 检查集合中是否存在指定的索引，复合索引
func CheckCollectionCompoundIndexExists(indexs []bson.M, fieldNames []string, checkUnique bool) bool {
	for _, index := range indexs {
		if keys, ok := index["key"].(bson.M); ok {
			if len(keys) == len(fieldNames) {
				allFieldsExist := true
				for _, fieldName := range fieldNames {
					if _, exists := keys[fieldName]; !exists {
						allFieldsExist = false
						break
					}
				}
				if allFieldsExist {
					if checkUnique {
						if unique, ok := index["unique"].(bool); ok && unique {
							return true
						}
					} else {
						return true
					}
				}
			}
		}
	}
	return false
}

type ModelInitFunc func(*MongoClient) error

var modelInitFuncs []ModelInitFunc

func AddModelInitFunc(f ModelInitFunc) {
	modelInitFuncs = append(modelInitFuncs, f)
}

func InitMongoDB() (*MongoClient, error) {
	mongoClient, err := NewMongoClient()
	if err != nil {
		log.Println("Error mongodb.NewMongoClient")
		return nil, err
	}

	InitModelBase(mongoClient) // 这个需要最先初始化

	for _, f := range modelInitFuncs {
		if err := f(mongoClient); err != nil {
			log.Println("Error modelInitFuncs")
			return nil, err
		}
	}

	return mongoClient, err
}
