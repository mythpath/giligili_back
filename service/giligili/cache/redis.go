package cache

import (
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/orm"
	"strconv"

	"github.com/go-redis/redis"
)

type RedisService struct {
	Config brick.Config           `inject:"config"`
	Orm    *orm.OrmService        `inject:"OrmService"`
	Logger *logging.LoggerService `inject:"LoggerService"`

	RedisClient *redis.Client
}

// Redis 在中间件中初始化redis链接
func (r *RedisService) Init() {
	redisDB:=r.Config.GetMapString("redis","redisDB","1")
	redisAddress:=r.Config.GetMapString("redis","address","127.0.0.1:6379")
	redisPassword:=r.Config.GetMapString("redis","password","")

	db, _ := strconv.ParseUint(redisDB, 10, 64)
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: redisPassword,
		DB:       int(db),
	})

	_, err := client.Ping().Result()

	if err != nil {
		panic(err)
	}

	r.RedisClient = client
}

// View 点击数
func (r *RedisService) View(ID uint) uint64 {
	countStr, _ := r.RedisClient.Get(VideoViewKey(ID)).Result()
	count, _ := strconv.ParseUint(countStr, 10, 64)

	return count
}
