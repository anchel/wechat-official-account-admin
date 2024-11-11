package mongodb

import (
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EntityWeixinMaterial struct {
	EntityBase `bson:",inline"`

	ID primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`

	AppID string `json:"appid" bson:"appid"`

	MediaCat  string `json:"media_cat" bson:"media_cat"` // temp-临时素材，perm-永久素材
	MediaType string `json:"media_type" bson:"media_type"`
	MediaId   string `json:"media_id" bson:"media_id"`

	FilePath    string `json:"file_path" bson:"file_path"`
	FileUrlPath string `json:"file_url_path" bson:"file_url_path"`

	WxUrl       string     `json:"wx_url" bson:"wx_url"`           // 微信侧的url，image类型会有
	Title       string     `json:"title" bson:"title"`             // video类型会有，临时素材没有
	Description string     `json:"description" bson:"description"` // video类型会有，临时素材没有
	ExpiresAt   *time.Time `json:"expires_at" bson:"expires_at"`   // 临时素材的过期时间
}

func (e *EntityWeixinMaterial) GetCreatedAt() time.Time {
	return e.CreatedAt
}

func (e *EntityWeixinMaterial) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

var ModelWeixinMaterial *ModelBase[EntityWeixinMaterial, *EntityWeixinMaterial]

func init() {

	AddModelInitFunc(func(client *MongoClient) error {
		log.Println("init mongodb model weixin downloadmedias")

		collectionName := "wx-materials"

		ModelWeixinMaterial = NewModelBase[EntityWeixinMaterial, *EntityWeixinMaterial](collectionName)

		// 检查索引是否存在

		return nil
	})
}
