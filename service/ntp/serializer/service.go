package serializer

import (
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/pkg/ntpServer/model"
)

type SerializerService struct {
	Config brick.Config           `inject:"config"`
	Orm    *orm.OrmService        `inject:"OrmService"`
	Logger *logging.LoggerService `inject:"LoggerService"`
}

// BuildVideo 序列化视频
func (s *SerializerService) BuildTimeServer(item model.TimeServer) TimeServer {
	var timeServer model.TimeServer
	if err := s.Orm.GetDB().Where("id = ?", item.ID).First(&timeServer).Error; err != nil {
		return TimeServer{}
	}

	return TimeServer{
		ID:      item.ID,
		URL:     item.URL,
		Comment: item.Comment,
	}
}

// BuildVideos 序列化视频列表
func (s *SerializerService) BuildTimeServers(items []model.TimeServer) (timeServers []TimeServer) {
	for _, item := range items {
		timeServer := s.BuildTimeServer(item)
		timeServers = append(timeServers, timeServer)
	}
	return
}
