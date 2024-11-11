package materialservice

import (
	"context"

	"github.com/anchel/wechat-official-account-admin/lib/types"
	"github.com/anchel/wechat-official-account-admin/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GetTempMaterialListResp struct {
	Total int64                           `json:"total"`
	List  []*mongodb.EntityWeixinMaterial `json:"list"`
}

// 获取临时素材列表
func GetLocalMaterialList(ctx context.Context, mediaCat string, mediaType string, offset, count int) (*GetTempMaterialListResp, error) {
	wxAppId := ctx.Value(types.ContextKey("appid"))
	findOptions := options.Find()
	findOptions.SetSkip(int64(offset))
	findOptions.SetLimit(int64(count))
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})
	filter := bson.D{{Key: "appid", Value: wxAppId}, {Key: "media_cat", Value: mediaCat}, {Key: "media_type", Value: mediaType}}

	// 获取总数量
	total, err := mongodb.ModelWeixinMaterial.Count(ctx, filter)
	if err != nil {
		return nil, err
	}

	docs, err := mongodb.ModelWeixinMaterial.FindMany(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	return &GetTempMaterialListResp{
		Total: total,
		List:  docs,
	}, nil
}
