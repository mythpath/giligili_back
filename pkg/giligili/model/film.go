package model

import (
	"gorm.io/gorm"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/service/common/gormHook"
)

// Film 影片模型
type Film struct {
	orm.SelfGormModel
	Name        string `json:"name"`
	Code        string `json:"code"`
	PerformerID uint   `json:"performer_id"`
	Post        string `gorm:"size:1000" json:"post"`
	Rank        string `json:"rank"`
	Comment     string `json:"comment"`
}

func FilmDesc() *orm.ModelDescriptor {
	return &orm.ModelDescriptor{
		Type: &Film{},
		New: func() interface{} {
			return &Film{}
		},
		NewSlice: func() interface{} {
			return &[]Film{}
		},
	}
}

func (f *Film) BeforeCreate(tx *gorm.DB) error {
	hook := gormHook.GetBeforeHook()
	if err := hook.CommonBeforeCreate(tx);err != nil {
		return err
	}

	return nil
}

func (f *Film) BeforeUpdate(tx *gorm.DB) error {
	hook:=gormHook.GetBeforeHook()
	if err := hook.CommonBeforeUpdate(tx); err != nil {
		return err
	}

	return nil
}
