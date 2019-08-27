package service

import (
	"selfText/giligili_back/model"
	"selfText/giligili_back/serializer"
)

// CreateVideoService 视频投稿的服务
type CreateVideoService struct {
	Title string `form:"title" json:"title" binding:"required,min=2,max=100"`
	Info  string `form:"info" json:"info" binding:"max=3000"`
	Url   string `form:"url" json:"url" binding:"required"`
	Avatar string `form:"avatar" json:"avatar"`
}

// Create 创建视频
func (service *CreateVideoService) Create(user *model.User) serializer.Response {
	video := model.Video{
		Title:  service.Title,
		Info:   service.Info,
		Url:    service.Url,
		Avatar: service.Avatar,
		UserID: user.ID,
	}

	err := model.DB.Create(&video).Error
	if err != nil {
		return serializer.Response{
			Status: 50001,
			Msg:    "视频保存失败",
			Error:  err.Error(),
		}
	}

	return serializer.Response{
		Data: serializer.BuildVideo(video),
	}
}
