package mongodb

import (
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EntityMenu struct {
	EntityBase `bson:",inline"`

	ID primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`

	AppID string `json:"appid" bson:"appid"`

	MenuType string `json:"menu_type" bson:"menu_type"` // normal, conditional
	MenuId   string `json:"menu_id" bson:"menu_id"`     // normal时是 normal，conditional时是 menuid
	MenuData string `json:"menu_data" bson:"menu_data"`
}

// 实现 ModelEntier 接口
func (e *EntityMenu) GetCreatedAt() time.Time {
	return e.CreatedAt
}

func (e *EntityMenu) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

var ModelMenu *ModelBase[EntityMenu, *EntityMenu]

func init() {

	AddModelInitFunc(func(client *MongoClient) error {
		log.Println("init mongodb model menu")

		collectionName := "wx-menus"

		ModelMenu = NewModelBase[EntityMenu, *EntityMenu](collectionName)

		// 检查索引是否存在

		return nil
	})
}
