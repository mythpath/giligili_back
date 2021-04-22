package sqlite

import (
	"fmt"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"path/filepath"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/orm"
)

// SQLite3Service
type SQLite3Service struct {
	orm.ModelRegistryImpl
	orm.DeleteCallback
	orm.UpdateCallback
	db     *gorm.DB
	Config brick.Config `inject:"config"`
}

func (p *SQLite3Service) Init() error {

	filename := p.Config.GetMapString("db", "url", "../data/nerv.db")
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); err != nil {
		if err := os.MkdirAll(dir, os.ModeDir|os.ModePerm); err != nil {
			return fmt.Errorf("create dir %s failed. %s", dir, err.Error())
		}
	}

	db, err := gorm.Open(sqlite.Open(filename), &gorm.Config{})
	if err != nil {
		return err
	}
	if sqlDB, dbErr := db.DB(); dbErr != nil {
		return dbErr
	} else {
		sqlDB.SetMaxOpenConns(1)
	}
	//db.DB().SetMaxOpenConns(1)

	//db.LogMode(false)
	//db.Logger.LogMode(logger.Info)
	for v := range p.Models() {
		db.AutoMigrate(v.Type)
	}
	if db.Error != nil {
		return db.Error
	}
	p.db = db
	// todo 隐去callback，如有问题再修正
	//p.DeleteCallback.Register(p, p.db)
	//p.UpdateCallback.Register(p, p.db)
	return nil
}

func (p *SQLite3Service) Dispose() error {
	//return p.db.Close()
	return nil
}

func (p *SQLite3Service) GetDB() *gorm.DB {
	return p.db
}
