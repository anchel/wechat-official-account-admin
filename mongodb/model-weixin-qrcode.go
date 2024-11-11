package mongodb

import (
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EntityWxQrcode struct {
	EntityBase `bson:",inline"`

	ID primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`

	AppID string `json:"appid" bson:"appid"`

	QrcodeType string `json:"qrcode_type" bson:"qrcode_type"` // temp-临时，limit-永久
	SceneStr   string `json:"scene_str" bson:"scene_str"`
	SceneId    int    `json:"scene_id" bson:"scene_id"`

	Title string `json:"title" bson:"title"`

	Ticket        string `json:"ticket" bson:"ticket"`
	ExpireSeconds int    `json:"expire_seconds" bson:"expire_seconds"`
	Url           string `json:"url" bson:"url"`
}

// 实现 ModelEntier 接口
func (e *EntityWxQrcode) GetCreatedAt() time.Time {
	return e.CreatedAt
}

func (e *EntityWxQrcode) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

var ModelWxQrcode *ModelBase[EntityWxQrcode, *EntityWxQrcode]

func init() {
	AddModelInitFunc(func(client *MongoClient) error {
		log.Println("init mongodb model user")

		collectionName := "wx-qrcodes"

		ModelWxQrcode = NewModelBase[EntityWxQrcode, *EntityWxQrcode](collectionName)

		return nil
	})
}
