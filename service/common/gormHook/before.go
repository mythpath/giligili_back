package gormHook

import (
	"gorm.io/gorm"
	"selfText/giligili_back/libcommon/orm"
)

// BeforeHook gorm通用更新前hook实现
type BeforeHook struct {
	defaultCreateFields []string
	defaultUpdateFields []string
}

// GetBeforeHook 获取默认实现的BeforeHook
func GetBeforeHook() BeforeHook {
	return BeforeHook{
		defaultCreateFields: []string{"CreatedBy", "UpdatedBy"},
		defaultUpdateFields: []string{"UpdatedBy"},
	}
}

// CommonBeforeCreate 在创建数据库记录前设置创建者和更新者字段
func (b *BeforeHook) CommonBeforeCreate(tx *gorm.DB) error {
	user := tx.Statement.Context.Value(orm.ContextCurrentUser())
	for _, field := range b.defaultCreateFields {
		if err := tx.Statement.Schema.LookUpField(field).Set(tx.Statement.ReflectValue, user); err != nil {
			return err
		}
	}

	return nil
}

// CommonBeforeUpdate 在更新数据库记录前设置更新者字段
func (b *BeforeHook) CommonBeforeUpdate(tx *gorm.DB) error {
	user := tx.Statement.Context.Value(orm.ContextCurrentUser())
	for _, field := range b.defaultUpdateFields {
		if err := tx.Statement.Schema.LookUpField(field).Set(tx.Statement.ReflectValue, user); err != nil {
			return err
		}
	}

	return nil
}
