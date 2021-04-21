package file

import (
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/libcommon/orm/dialects/mysql"
	"testing"
)

func TestGetFilesAndDirs(t *testing.T) {
	c := brick.NewContainer()
	c.Add(&brick.JSONConfigService{}, "config", brick.FactoryFunc(func() interface{} {
		return brick.NewJSONConfigService("/Users/liyuxin/go/src/selfText/giligili_back/cmd/ntpServer/config/config.json")
	}))

	c.Add(&orm.OrmService{}, "OrmService", nil)
	c.Add(&mysql.MySQLService{}, "DB", nil)
	c.Add(&logging.LoggerService{}, "LoggerService", nil)

	c.Add(&FileServer{}, "FileServer", nil)

	c.Build()

	dirPath := "/Users/liyuxin/Documents/工作/网太/文档"
	fileServer := c.GetByName("FileServer").(*FileServer)
	files, dirs, err := fileServer.GetFilesAndDirs(dirPath, "", 1)
	if err != nil {
		t.Error(err)
	}
	t.Log("print files:")
	for _, file := range files {
		t.Log(file)
	}
	t.Log("print dirs:")
	for _, dir := range dirs {
		t.Log(dir)
	}
}
