package mysql

import (
	//"database/sql"
	"fmt"
	"gorm.io/gorm/logger"

	//"github.com/go-sql-driver/mysql"
	"selfText/giligili_back/libcommon/logging"
	"time"

	"log"

	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/codec"
	"selfText/giligili_back/libcommon/orm"

	//"github.com/jinzhu/gorm"
	//_ "github.com/jinzhu/gorm/dialects/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	key string = "Jdb@Mysql123"
)

// MySQLService
type MySQLService struct {
	orm.ModelRegistryImpl
	orm.CreateCallback
	orm.DeleteCallback
	orm.UpdateCallback
	db     *gorm.DB
	Config brick.Config           `inject:"config"`
	Logger *logging.LoggerService `inject:"LoggerService"`
}

func (p *MySQLService) Init() error {
	//fmt.Println("Init db")

	passwd := p.Config.GetMapString("db", "password", "root")
	encrypt := p.Config.GetMapBool("db", "encrypt", true)
	var err error
	if encrypt {
		passwd, err = p.descrypt(passwd)
		if err != nil {
			return err
		}
	}

	url := fmt.Sprintf(
		"%s:%s@%s",
		p.Config.GetMapString("db", "user", "root"),
		passwd,
		p.Config.GetMapString("db", "url"),
	)
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{})
	if err != nil {
		return err
	}

	timeout, err := time.ParseDuration(p.Config.GetMapString("db", "connMaxLifetime", "2h"))
	if err != nil {
		timeout = 2 * time.Hour
	}
	if sqlDB, dbErr := db.DB(); dbErr != nil {
		sqlDB.SetConnMaxLifetime(timeout)
		sqlDB.SetMaxOpenConns(p.Config.GetMapInt("db", "maxOpenConns", 12))
		sqlDB.SetMaxIdleConns(p.Config.GetMapInt("db", "maxIdleConns", 12))
	}
	//db.DB().SetConnMaxLifetime(timeout)
	//
	//db.DB().SetMaxOpenConns(p.Config.GetMapInt("db", "maxOpenConns", 12))
	//
	//db.DB().SetMaxIdleConns(p.Config.GetMapInt("db", "maxIdleConns", 12))

	log := p.Config.GetMapBool("db", "log", false)
	//db.LogMode(log)
	//db.SetLogger(p.Logger)
	if log {
		db.Logger = logger.Default.LogMode(logger.Info)
	}
	// disable it if you think it's ntp-consuming
	if p.Config.GetMapBool("db", "autoMigrate", true) {
		for v := range p.Models() {
			if err := db.AutoMigrate(v.Type); err != nil {
				return err
			}
		}
	}

	p.db = db
	// todo 隐去callback，如有问题再修正
	//p.CreateCallback.Register(p, p.db)
	//p.DeleteCallback.Register(p, p.db)
	//p.UpdateCallback.Register(p, p.db)
	return nil
}

func (p *MySQLService) Dispose() error {
	//if p.db != nil {
	//	return p.db.Close()
	//} else {
	//	return nil
	//}
	return nil
}

func (p *MySQLService) GetDB() *gorm.DB {
	return p.db
}

func (p *MySQLService) descrypt(data string) (passwd string, err error) {
	aesCipher := codec.NewAESCipher(codec.AES128, key)
	if err = aesCipher.Init(); err != nil {
		log.Printf("mysql init error. err:%v", err)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("descrypt panic. err:%v", err)
			err = fmt.Errorf("%v", r)
		}
	}()

	passwd = string(aesCipher.Decrypt(data))

	return
}
