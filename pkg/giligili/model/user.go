package model

import (
	"gorm.io/gorm"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/service/common/gormHook"
)

// User 用户模型
type User struct {
	orm.SelfGormModel
	UserName       string `json:"user_name"`
	PasswordDigest string `json:"password_digest"`
	Nickname       string `json:"nickname"`
	Status         string `json:"status"`
	Avatar         string `gorm:"size:1000" json:"avatar"`
}

func UserDesc() *orm.ModelDescriptor {
	return &orm.ModelDescriptor{
		Type: &User{},
		New: func() interface{} {
			return &User{}
		},
		NewSlice: func() interface{} {
			return &[]User{}
		},
	}
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	hook := gormHook.GetBeforeHook()
	if err := hook.CommonBeforeCreate(tx);err != nil {
		return err
	}

	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	hook:=gormHook.GetBeforeHook()
	if err := hook.CommonBeforeUpdate(tx); err != nil {
		return err
	}

	return nil
}
