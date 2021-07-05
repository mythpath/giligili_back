package main

import (
	"flag"
	"selfText/giligili_back/cmd/fileServer/service"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/net/http/invoker"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/libcommon/orm/dialects/mysql"
	"selfText/giligili_back/libcommon/passport/auth"
	"selfText/giligili_back/pkg/fileServer/model"
	"selfText/giligili_back/service/file/file"
	"selfText/giligili_back/service/file/serializer"
	"selfText/giligili_back/service/file/snapshot"
)

func main() {
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

	container.Add(&serializer.SerializerService{}, "SerializerService", nil)
	container.Add(&file.UploadServer{}, "UploadServer", nil)
	container.Add(&snapshot.SnapshotServer{}, "SnapshotServer", nil)
	container.Add(&file.FileRouting{}, "FileServer", nil)
	container.Add(&snapshot.SnapshotServer{}, "SnapshotServer", nil)

	container.Build()

	select {}
}
