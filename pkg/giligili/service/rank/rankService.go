package rank

import (
	"context"
	"fmt"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/pkg/giligili/model"
	"selfText/giligili_back/service/giligili/cache"
	"selfText/giligili_back/service/giligili/serializer"
	"strings"
)

// DailyRankService 每日排行服务
type DailyRankService struct {
	Config brick.Config           `inject:"config"`
	Orm    *orm.OrmService        `inject:"OrmService"`
	Logger *logging.LoggerService `inject:"LoggerService"`

	SerializerService *serializer.SerializerService `inject:"SerializerService"`
	RedisService      *cache.RedisService           `inject:"RedisService"`
}

// Get 获取排行
func (d *DailyRankService) Get(ctx context.Context) (serializer.Response, error) {
	var videos []model.Video

	// 从redis中读取点击数前十的视频
	vids, err := d.RedisService.RedisClient.ZRevRange(cache.DailyRankKey, 0, 9).Result()
	if len(vids) < 1 {
		return serializer.Response{
			Status: 50000,
			Data:   nil,
			Msg:    "Redis中没有对应数据",
			Error:  err.Error(),
		}, err
	}

	order := fmt.Sprintf("FIELD(id, %s)", strings.Join(vids, ","))
	if err := d.Orm.GetDB().Where("id in (?)", vids).Order(order).Find(&videos).Error; err != nil {
		return serializer.Response{
			Status: 50000,
			Data:   nil,
			Msg:    "数据库连接错误",
			Error:  err.Error(),
		}, err
	}

	return serializer.Response{
		Data: d.SerializerService.BuildVideos(videos),
	}, nil
}
