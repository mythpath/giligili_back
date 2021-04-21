package model

import (
	"gorm.io/gorm"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/service/common/gormHook"
)

type TimeServer struct {
	orm.SelfGormModel
	URL     string `json:"url"`
	Comment string `json:"comment"`
}

func TimeServerDesc() *orm.ModelDescriptor {
	return &orm.ModelDescriptor{
		Type: &TimeServer{},
		New: func() interface{} {
			return &TimeServer{}
		},
		NewSlice: func() interface{} {
			return &[]TimeServer{}
		},
	}
}

func (t *TimeServer) BeforeCreate(tx *gorm.DB) error {
	hook := gormHook.GetBeforeHook()
	if err := hook.CommonBeforeCreate(tx);err != nil {
		return err
	}

	return nil
}

func (t *TimeServer) BeforeUpdate(tx *gorm.DB) error {
	hook:=gormHook.GetBeforeHook()
	if err := hook.CommonBeforeUpdate(tx); err != nil {
		return err
	}

	return nil
}
