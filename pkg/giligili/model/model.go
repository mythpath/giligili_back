package model

import (
	//
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"selfText/giligili_back/libcommon/orm"
)

var (
	FilmM      = "Film"
	PerformerM = "Performer"
	UserM      = "User"
	VideoM     = "Video"
)

type Model struct {
	Register orm.ModelRegistry `inject:"DB"`
}

func (m *Model) AfterNew() {
	m.Register.Put(FilmM, FilmDesc())
	m.Register.Put(PerformerM, PerformerDesc())
	m.Register.Put(UserM, UserDesc())
	m.Register.Put(VideoM, VideoDesc())
}
