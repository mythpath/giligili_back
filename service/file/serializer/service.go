package serializer

import (
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/orm"
)

type SerializerService struct {
	Config brick.Config           `inject:"config"`
	Orm    *orm.OrmService        `inject:"OrmService"`
	Logger *logging.LoggerService `inject:"LoggerService"`
}

//// BuildTimeServer 序列化时间服务器
//func (s *SerializerService) BuildTimeServer(item model.TimeServer) TimeServer {
//	var timeServer model.TimeServer
//	if err := s.Orm.GetDB().Where("id = ?", item.ID).First(&timeServer).Error; err != nil {
//		return TimeServer{}
//	}
//
//	return TimeServer{
//		ID:      item.ID,
//		URL:     item.URL,
//		Comment: item.Comment,
//	}
//}
//
//// BuildTimeServers 序列化时间服务器列表
//func (s *SerializerService) BuildTimeServers(items []model.TimeServer) (timeServers []TimeServer) {
//	for _, item := range items {
//		timeServer := s.BuildTimeServer(item)
//		timeServers = append(timeServers, timeServer)
//	}
//	return
//}
