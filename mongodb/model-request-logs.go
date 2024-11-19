package mongodb

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EntityRequestLog struct {
	EntityBase `bson:",inline"`

	ID primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`

	AppID string `json:"appid" bson:"appid"`

	Time      time.Time `json:"time" bson:"time"`
	Status    int       `json:"status" bson:"status"`
	Latency   float64   `json:"latency" bson:"latency"`
	Ip        string    `json:"ip" bson:"ip"`
	Method    string    `json:"method" bson:"method"`
	Path      string    `json:"path" bson:"path"`
	Query     string    `json:"query" bson:"query"`
	Body      string    `json:"body" bson:"body"`
	UserAgent string    `json:"user-agent" bson:"user-agent"`
}

// 实现 ModelEntier 接口
func (e *EntityRequestLog) GetCreatedAt() time.Time {
	return e.CreatedAt
}

func (e *EntityRequestLog) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

var ModelRequestLog *ModelBase[EntityRequestLog, *EntityRequestLog]

func init() {
	AddModelInitFunc(func(client *MongoClient) error {
		log.Println("init mongodb model request-logs")

		collectionName := "request-logs"

		ModelRequestLog = NewModelBase[EntityRequestLog, *EntityRequestLog](collectionName)

		// 检查索引是否存在
		collection, err := mongoClient.GetCollection(collectionName)
		if err != nil {
			log.Println("Error mongoClient.GetCollection")
			return err
		}
		usersIndexs, err := GetCollectionIndexs(context.Background(), collection)
		if err != nil {
			log.Println("Error GetCollectionIndexs")
			return err
		}
		if !CheckCollectionCompoundIndexExists(usersIndexs, []string{"appid", "time"}, false) {
			_, err = collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
				Keys: bson.D{
					{Key: "appid", Value: 1},
					{Key: "time", Value: 1},
				},
				Options: options.Index().SetUnique(false),
			})
			if err != nil {
				log.Println("Error CreateOne")
				return err
			}
		}

		return nil
	})
}
