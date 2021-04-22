package orm

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// CreateCallback
type CreateCallback struct {
	models ModelRegistry
}

func (p *CreateCallback) Register(models ModelRegistry, db *gorm.DB) {
	p.models = models
	//gorm.DefaultCallback.Create().Replace("gorm:update_time_stamp", p.updateTimeStampForCreateCallback)
	db.Callback().Create().Replace("gorm:update_time_stamp", p.updateTimestampForCreateCallbackV2)
}

// updateTimeStampForCreateCallback write for gorm v1
//func (p *CreateCallback) updateTimeStampForCreateCallback(scope *gorm.Scope) {
//if !scope.HasError() {
//	//fmt.Printf("updateTimeStampForCreateCallback: %s\n", scope.GetModelStruct().ModelType.Name())
//	now := NowFunc()
//	scope.SetColumn("CreatedAt", now)
//	scope.SetColumn("UpdatedAt", now)
//
//	ctx := scope.GetContext()
//	v := ctx.Value(gorm.ContextCurrentUser())
//	switch user := v.(type) {
//	case string:
//		if field, ok := scope.FieldByName("CreatedBy"); ok {
//			field.Set(user)
//		}
//		if field, ok := scope.FieldByName("UpdatedBy"); ok {
//			field.Set(user)
//		}
//	}
//}
//}

func (p *CreateCallback) updateTimestampForCreateCallbackV2(tx *gorm.DB) {
	if tx.Statement.Schema != nil && tx.Statement.Error == nil {
		// 裁剪图片字段并上传到CDN，dummy code
		now := NowFunc()
		// 使用字段名或数据库名查找字段
		fieldNames := []string{"CreatedAt", "UpdatedAt", "CreatedBy", "UpdatedBy"}
		for _, fieldName := range fieldNames[:2] {
			field := tx.Statement.Schema.LookUpField(fieldName)
			if err := field.Set(tx.Statement.ReflectValue, now); err != nil {
				logrus.Fatalf("modify %s error in callback. Error: %s", fieldName, err)
			}
			//if err := tx.Statement.Schema.LookUpField(fieldName).Set(tx.Statement.ReflectValue, now); err != nil {
			//	logrus.Fatalf("modify %s error in callback. Error: %s", field, err)
			//}
		}
		//ctx := tx.Statement.Context
		v := tx.Statement.Context.Value(ContextCurrentUser())
		switch user := v.(type) {
		case string:
			for _, field := range fieldNames[2:] {
				if err := tx.Statement.Schema.LookUpField(field).Set(tx.Statement.ReflectValue, user); err != nil {
					logrus.Fatalf("modify %s error in callback. Error: %s", field, err)
				}
			}
		}
	}
}
