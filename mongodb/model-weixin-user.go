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

type EntityWeixinUser struct {
	EntityBase `bson:",inline"`

	ID primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`

	AppID string `json:"appid" bson:"appid"`

	OpenID         string     `json:"openid" bson:"openid"`
	UnionID        string     `json:"unionid" bson:"unionid"`
	Nickname       string     `json:"nickname" bson:"nickname"`
	Avatar         string     `json:"avatar" bson:"avatar"`
	Subscribed     bool       `json:"subscribed" bson:"subscribed"`
	SubscribedAt   *time.Time `json:"subscribed_at,omitempty" bson:"subscribed_at,omitempty"`
	UnSubscribedAt *time.Time `json:"unsubscribed_at,omitempty" bson:"unsubscribed_at,omitempty"`

	SceneID string `json:"scene_id" bson:"scene_id"`
}

// 实现 ModelEntier 接口
func (e *EntityWeixinUser) GetCreatedAt() time.Time {
	return e.CreatedAt
}

func (e *EntityWeixinUser) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

var ModelWeixinUser *ModelBase[EntityWeixinUser, *EntityWeixinUser]

func init() {
	AddModelInitFunc(func(client *MongoClient) error {
		log.Println("init mongodb model weixin user")

		collectionName := "wx-users"

		ModelWeixinUser = NewModelBase[EntityWeixinUser, *EntityWeixinUser](collectionName)

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
		if !CheckCollectionIndexExists(context.Background(), usersIndexs, "openid", true) {
			_, err = collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
				Keys: bson.M{
					"openid": 1,
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
