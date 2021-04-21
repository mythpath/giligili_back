package serializer

import (
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/pkg/giligili/model"
	"selfText/giligili_back/service/giligili/cache"
	"selfText/giligili_back/service/giligili/ossService"
)

type SerializerService struct {
	Config brick.Config           `inject:"config"`
	Orm    *orm.OrmService        `inject:"OrmService"`
	Logger *logging.LoggerService `inject:"LoggerService"`

	RedisService *cache.RedisService    `inject:"RedisService"`
	OSSService   *ossService.OSSServcie `inject:"OSSService"`
}

// BuildVideo 序列化视频
func (s *SerializerService) BuildVideo(item model.Video) Video {
	var user model.User
	if err := s.Orm.GetDB().Where("id = ?", item.UserID).First(&user).Error; err != nil {
		return Video{}
	}

	return Video{
		ID:        item.ID,
		Title:     item.Title,
		Info:      item.Info,
		Url:       item.Url,
		Avatar:    s.OSSService.AvatarURL(item.ID),
		View:      s.RedisService.View(item.ID),
		User:      s.BuildUser(user),
		CreatedAt: item.CreatedAt.Unix(),
	}
}

// BuildVideos 序列化视频列表
func (s *SerializerService) BuildVideos(items []model.Video) (videos []Video) {
	for _, item := range items {
		video := s.BuildVideo(item)
		videos = append(videos, video)
	}
	return videos
}

// BuildUser 序列化用户
func (s *SerializerService) BuildUser(user model.User) User {
	return User{
		ID:        user.ID,
		UserName:  user.UserName,
		Nickname:  user.Nickname,
		Status:    user.Status,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt.Unix(),
	}
}

// BuildUserResponse 序列化用户响应
func (s *SerializerService) BuildUserResponse(user model.User) UserResponse {
	return UserResponse{
		Data: s.BuildUser(user),
	}
}

// BuildPerformer 序列化演员
func (s *SerializerService) BuildPerformer(performer model.Performer) Performer {
	return Performer{
		Name:    performer.Name,
		Chinese: performer.Chinese,
		Gender:  performer.Gender,
		Locate:  performer.Locate,
		Avatar:  performer.Avatar,
		Rank:    performer.Rank,
		Comment: performer.Comment,
	}
}

// BuildPerformers 序列化演员列表
func (s *SerializerService) BuildPerformers(items []model.Performer) (performers []Performer) {
	for _, item := range items {
		performer := s.BuildPerformer(item)
		performers = append(performers, performer)
	}
	return performers
}

// BuildFilm 序列化影片
func (s *SerializerService) BuildFilm(film model.Film) Film {
	return Film{
		Name:        film.Name,
		Code:        film.Code,
		PerformerID: film.PerformerID,
		Post:        film.Post,
		Rank:        film.Rank,
		Comment:     film.Comment,
	}
}

// BuildFilms 序列化影片列表
func (s *SerializerService) BuildFilms(items []model.Film) (films []Film) {
	for _, item := range items {
		film := s.BuildFilm(item)
		films = append(films, film)
	}
	return films
}

// BuildListResponse 列表构建器
func (s *SerializerService) BuildListResponse(items interface{}, total uint) Response {
	return Response{
		Data: DataList{
			Items: items,
			Total: total,
		},
	}
}
