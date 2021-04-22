package lock_test

import (
	"testing"

	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/libcommon/parallel/lock"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/stretchr/testify/assert"
)

func TestLock_TryLock(t *testing.T) {

	initDB()

	lock1 := lock.GetLock("obj", 1)
	lock2 := lock.GetLock("obj", 1)
	defer lock1.Unlock()
	defer lock2.Unlock()

	ok1 := lock1.TryLock()
	assert.Equal(t, ok1, true)

	ok2 := lock2.TryLock()
	assert.Equal(t, ok2, false)
}

func initDB() {
	gdb, err := gorm.Open("mysql", "root:root@/nerv?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	orm.DB = gdb
	orm.DB.LogMode(true)
	for _, v := range orm.Models {
		orm.DB.AutoMigrate(v.Type)
	}
}
