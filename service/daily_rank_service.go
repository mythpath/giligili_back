package service

import (
	"fmt"
	"selfText/giligili_back/cache"
	"selfText/giligili_back/model"
	"selfText/giligili_back/serializer"
	"strings"
)

// DailyRankService 每日排行服务
type DailyRankService struct {
}

// Get 获取排行
func (service *DailyRankService) Get() serializer.Response {
	var videos []model.Video

	// 从redis中读取点击数前十的视频
	vids, _ := cache.RedisClient.ZRevRange(cache.DailyRankKey, 0, 9).Result()
	if len(vids)<1{
		return serializer.Response{
			Status: 50000,
			Data:   nil,
			Msg:    "Redis中没有对应数据",
			Error:  "",
		}
	}

	order := fmt.Sprintf("FIELD(id, %s)", strings.Join(vids, ","))
	err := model.DB.Where("id in (?)", vids).Order(order).Find(&videos).Error
	if err != nil {
		return serializer.Response{
			Status: 50000,
			Data:   nil,
			Msg:    "数据库连接错误",
			Error:  err.Error(),
		}
	}

	return serializer.Response{
		Data: serializer.BuildVideos(videos),
	}
}
