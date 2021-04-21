package model

import (
	"gorm.io/gorm"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/service/common/gormHook"
)

// Video 视频模型
type Video struct {
	orm.SelfGormModel
	Title  string `json:"title"`
	Info   string `gorm:"type:text" json:"info"`
	Url    string `json:"url"`
	Avatar string `json:"avatar"`
	UserID uint `json:"user_id"`
}

func VideoDesc() *orm.ModelDescriptor {
	return &orm.ModelDescriptor{
		Type: &Video{},
		New: func() interface{} {
			return &Video{}
		},
		NewSlice: func() interface{} {
			return &[]Video{}
		},
	}
}

func (v *Video) BeforeCreate(tx *gorm.DB) error {
	hook := gormHook.GetBeforeHook()
	if err := hook.CommonBeforeCreate(tx);err != nil {
		return err
	}

	return nil
}

func (v *Video) BeforeUpdate(tx *gorm.DB) error {
	hook:=gormHook.GetBeforeHook()
	if err := hook.CommonBeforeUpdate(tx); err != nil {
		return err
	}

	return nil
}
