package main

import (
	"flag"
	"selfText/giligili_back/cmd/giligili/service"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/metrics"
	"selfText/giligili_back/libcommon/net/http/invoker"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/libcommon/orm/dialects/mysql"
	"selfText/giligili_back/libcommon/passport/auth"
	"selfText/giligili_back/pkg/giligili/model"
	"selfText/giligili_back/pkg/giligili/service/film"
	"selfText/giligili_back/pkg/giligili/service/performer"
	"selfText/giligili_back/pkg/giligili/service/rank"
	"selfText/giligili_back/pkg/giligili/service/upload"
	"selfText/giligili_back/pkg/giligili/service/user"
	"selfText/giligili_back/pkg/giligili/service/video"
	"selfText/giligili_back/service/giligili/cache"
	"selfText/giligili_back/service/giligili/metric"
	"selfText/giligili_back/service/giligili/ossService"
	"selfText/giligili_back/service/giligili/serializer"
	"selfText/giligili_back/service/giligili/tasks"
)

func main() {
	//// 从配置文件读取配置
	//conf.Init()
	//
	//// 装载路由
	//r := server.NewRouter()
	//r.Run(":3000")
	configPath := flag.String("c", "../config/config.json", "configuration file")
	flag.Parse()

	container := brick.NewContainer()
	container.Add(&brick.JSONConfigService{}, "config", brick.FactoryFunc(func() interface{} {
		return brick.NewJSONConfigService(*configPath)
	}))
	container.Add(&orm.OrmService{}, "OrmService", nil)
	container.Add(&invoker.Invoker{}, "Invoker", nil)
	container.Add(&mysql.MySQLService{}, "DB", nil)
	container.Add(&logging.LoggerService{}, "LoggerService", nil)
	container.Add(&service.Http{}, "Http", nil)
	container.Add(&model.Model{}, "Model", nil)
	container.Add(&auth.PassportClient{}, "passport-client", nil)

	container.Add(&film.FilmService{}, "FilmService", nil)
	container.Add(&performer.PerformerService{}, "PerformerService", nil)
	container.Add(&rank.DailyRankService{}, "DailyRankService", nil)
	container.Add(&upload.OSSTokenService{}, "OSSTokenService", nil)
	container.Add(&user.UserService{}, "UserService", nil)
	container.Add(&user.UserLoginService{}, "UserLoginService", nil)
	container.Add(&user.UserRegisterService{}, "UserRegisterService", nil)
	container.Add(&video.VideoService{}, "VideoService", nil)

	container.Add(&cache.RedisService{}, "RedisService", nil)
	container.Add(&ossService.OSSServcie{}, "OSSService", nil)
	container.Add(&serializer.SerializerService{}, "SerializerService", nil)
	container.Add(&tasks.CronService{}, "CronService", nil)
	container.Add(&tasks.RankService{}, "RankService", nil)

	// metric
	//container.Add(&cl.Collector{}, "Collector", nil)
	container.Add(&metric.InternalCollector{}, "MetricsInternalCollector", nil)
	container.Add(&metrics.CollectorService{}, "MetricsCollector", nil)
	container.Add(&metrics.RegistryService{}, "MetricsRegistry", nil)
	container.Add(&metrics.HttpExporterService{}, "MetricsHttpExporter", nil)

	container.Build()

	select {}
}
