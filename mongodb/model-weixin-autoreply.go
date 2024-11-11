package mongodb

import (
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EntityWeixinAutoReply struct {
	EntityBase `bson:",inline"`

	ID primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`

	AppID string `json:"appid" bson:"appid"`

	ReplyType string `json:"reply_type" bson:"reply_type"` // subscribe, keyword, message, menu_click
	ReplyData string `json:"reply_data" bson:"reply_data"`

	ExtId string `json:"ext_id" bson:"ext_id"` // 暂时是reply_type=menu_click时有效。对应的本地数据库的菜单ID

	DraftData string `json:"draft_data" bson:"draft_data"` // 草稿数据，暂时是菜单点击回复时，未发布到微信时的数据

	RuleTitle   string   `json:"rule_title" bson:"rule_title"`     // reply_type=keyword时有效
	Keywords    []string `json:"keywords" bson:"keywords"`         // reply_type=keyword时有效
	KeywordsDef string   `json:"keywords_def" bson:"keywords_def"` // reply_type=keyword时有效
}

func (e *EntityWeixinAutoReply) GetCreatedAt() time.Time {
	return e.CreatedAt
}

func (e *EntityWeixinAutoReply) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

var ModelWeixinAutoReply *ModelBase[EntityWeixinAutoReply, *EntityWeixinAutoReply]

func init() {

	AddModelInitFunc(func(client *MongoClient) error {
		log.Println("init mongodb model weixin autoreply")

		collectionName := "wx-autoreply"

		ModelWeixinAutoReply = NewModelBase[EntityWeixinAutoReply, *EntityWeixinAutoReply](collectionName)

		// 检查索引是否存在

		return nil
	})
}
