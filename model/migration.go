package model

//执行数据迁移

func migration() {
	// 自动迁移模式
	DB.Set("grom:table_options", "charset=utf8mb4").
		AutoMigrate(&User{}).
		AutoMigrate(&Video{})
}
