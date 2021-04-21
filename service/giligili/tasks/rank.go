package tasks

import (
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/service/giligili/cache"
)

type RankService struct {
	Config brick.Config           `inject:"config"`
	Orm    *orm.OrmService        `inject:"OrmService"`
	Logger *logging.LoggerService `inject:"LoggerService"`

	RedisService *cache.RedisService `inject:"RedisService"`
}

// RestartDailyRank 重启一天的排名
func (r *RankService)RestartDailyRank() error {
	return r.RedisService.RedisClient.Del("rank:daily").Err()
}
