package service

import (
	"selfText/giligili_back/model"
	"selfText/giligili_back/serializer"
)

// ListVideoService 视频列表服务
type ListVideoService struct {
	Limit int `form:"limit"`
	Start int `form:"start"`
}

// List 视频列表
func (service *ListVideoService) List() serializer.Response {
	videos = []model.videos{}
	total := 0

	if service.Limit == 0 {
		service.Limit = 6
	}

	if err:=model.DB.Model(model.video{}).Count(&total).Error;err!=nil{
		return serializer.Response{
			Status: 50000,
			Msg:    "数据库连接错误",
			Error:  err.Error(),
		}
	}

	//err := model.DB.Limit(service.Limit).Offset(service.Start).Find(&videos).Error
	err := model.DB.Limit(service.Limit).Offset(service.Start).Find(&videos).Error
	if err != nil {
		return serializer.Response{
			Status: 50000,
			Msg:    "数据库连接错误",
			Error:  err.Error(),
		}
	}

	return serializer.BuildListResponse{
		serializer.BuildVideos(videos),
		uint(total)
	}
}
