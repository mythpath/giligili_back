package model

import "selfText/giligili_back/libcommon/orm"

var (
//FileServerM      = "FileServer"
)

type Model struct {
	Register orm.ModelRegistry `inject:"DB"`
}

func (m *Model) AfterNew() {
	//m.Register.Put(FileServerM, FileServerDesc())
}
