package model

import (
	"gorm.io/gorm"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/service/common/gormHook"
)

// performer 演员模型
type Performer struct {
	orm.SelfGormModel
	Name    string `json:"name"`
	Chinese string `json:"chinese"`
	Gender  int `binding:"required min=0,max=1" json:"gender"`
	Locate  string `json:"locate"`
	Avatar  string `gorm:"size:1000" json:"avatar"`
	Rank    string `json:"rank"`
	Comment string `json:"comment"`
}

func PerformerDesc() *orm.ModelDescriptor {
	return &orm.ModelDescriptor{
		Type: &Performer{},
		New: func() interface{} {
			return &Performer{}
		},
		NewSlice: func() interface{} {
			return &[]Performer{}
		},
	}
}

func (p *Performer) BeforeCreate(tx *gorm.DB) error {
	hook := gormHook.GetBeforeHook()
	if err := hook.CommonBeforeCreate(tx);err != nil {
		return err
	}

	return nil
}

func (p *Performer) BeforeUpdate(tx *gorm.DB) error {
	hook:=gormHook.GetBeforeHook()
	if err := hook.CommonBeforeUpdate(tx); err != nil {
		return err
	}

	return nil
}
