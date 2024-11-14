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

type EntityWxAppid struct {
	EntityBase `bson:",inline"`

	ID primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`

	AppType string `json:"app_type" bson:"app_type"` // 订阅号，公众号，小程序
	Name    string `json:"name" bson:"name"`

	AppID          string `json:"appid" bson:"appid"`
	AppSecret      string `json:"appsecret" bson:"appsecret"`
	Token          string `json:"token" bson:"token"`
	EncodingAESKey string `json:"encoding_aes_key" bson:"encoding_aes_key"`

	Thumbnail string `json:"thumbnail" bson:"thumbnail"` // 缩略图

	EnabledAutoReplyKeyword   bool `json:"enabled_auto_reply_keyword" bson:"enabled_auto_reply_keyword"`     // 是否启用关键词回复
	EnabledAutoReplyMessage   bool `json:"enabled_auto_reply_message" bson:"enabled_auto_reply_message"`     // 是否启用消息回复
	EnabledAutoReplySubscribe bool `json:"enabled_auto_reply_subscribe" bson:"enabled_auto_reply_subscribe"` // 是否启用关注回复
}

// 实现 ModelEntier 接口
func (e *EntityWxAppid) GetCreatedAt() time.Time {
	return e.CreatedAt
}

func (e *EntityWxAppid) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

var ModelWxAppid *ModelBase[EntityWxAppid, *EntityWxAppid]

func init() {
	AddModelInitFunc(func(client *MongoClient) error {
		log.Println("init mongodb model user")

		collectionName := "wx-appids"

		ModelWxAppid = NewModelBase[EntityWxAppid, *EntityWxAppid](collectionName)

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
		if !CheckCollectionIndexExists(usersIndexs, "appid", true) {
			_, err = collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
				Keys: bson.M{
					"appid": 1,
				},
				Options: options.Index().SetUnique(true),
			})
			if err != nil {
				log.Println("Error CreateOne")
				return err
			}
		}

		return nil
	})
}
