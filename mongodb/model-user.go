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

type EntityUser struct {
	EntityBase `bson:",inline"`

	ID primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`

	UserType string `json:"user_type" bson:"user_type"`

	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
	Email    string `json:"email" bson:"email"`
	Remark   string `json:"remark" bson:"remark"`
}

// 实现 ModelEntier 接口
func (e *EntityUser) GetCreatedAt() time.Time {
	return e.CreatedAt
}

func (e *EntityUser) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

var ModelUser *ModelBase[EntityUser, *EntityUser]

func init() {
	AddModelInitFunc(func(client *MongoClient) error {
		log.Println("init mongodb model user")

		collectionName := "users"

		ModelUser = NewModelBase[EntityUser, *EntityUser](collectionName)

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
		if !CheckCollectionIndexExists(context.Background(), usersIndexs, "username", true) {
			_, err = collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
				Keys: bson.M{
					"username": 1,
				},
				Options: options.Index().SetUnique(true),
			})
			if err != nil {
				log.Println("Error CreateOne")
				return err
			}
		}
		if !CheckCollectionIndexExists(context.Background(), usersIndexs, "email", true) {
			_, err = collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
				Keys: bson.M{
					"email": 1,
				},
				Options: options.Index().SetUnique(true),
			})
			if err != nil {
				log.Println("Error CreateOne")
				return err
			}
		}

		ctx := context.TODO()
		// 判断初始用户 admin 是否存在，不存在则创建
		filter := bson.M{"username": "admin"}
		var doc EntityUser
		err = collection.FindOne(ctx, filter).Decode(&doc)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				// 不存在
				doc = EntityUser{}
				doc.Username = "admin"
				doc.Password = "31e9fb146377ca1ec73f07bf68382acb" // admin1987
				doc.Email = "admin@qq.com"
				doc.UserType = "admin"
				doc.Remark = "初始管理员用户"
				doc.CreatedAt = time.Now()
				_, err = collection.InsertOne(ctx, doc)
				if err != nil {
					log.Println("创建用户 admin 错误")
					return err
				}
				log.Println("创建用户 admin 成功")
			} else {
				log.Println("查找用户 admin 错误")
				return err
			}
		}

		return nil
	})
}
