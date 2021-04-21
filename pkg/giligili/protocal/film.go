package protocal

import "selfText/giligili_back/libcommon/nerror"

// FilmCreateInput 电影信息创建输入结构
type FilmCreateInput struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	PerformerID uint   `json:"performer_id"`
	Post        string `gorm:"size:1000" json:"post"`
	Rank        string `json:"rank"`
	Comment     string `json:"comment"`
}

func (f *FilmCreateInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if f.Name == "" && f.Code=="" {
		return nerr.FieldError("name", "One of the name and code must exist")
	}

	return nil
}

// FilmDeleteInput 电影信息删除输入结构
type FilmDeleteInput struct {
	ID uint `json:"id" invoke-path:"id"`
}

func (f *FilmDeleteInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if f.ID <= 0 {
		return nerr.FieldError("id", "id can not be lower than 0")
	}

	return nil
}

// FilmUpdateInput 电影信息更新输入结构
type FilmUpdateInput struct {
	ID uint `json:"id" invoke-path:"id"`
	FilmCreateInput
}

func (f *FilmUpdateInput) Validate() error {
	if err := f.FilmCreateInput.Validate(); err != nil {
		return err
	}
	nerr := nerror.NewArgumentError("")

	if f.ID <= 0 {
		return nerr.FieldError("id", "id can not be lower than 0")
	}

	return nil
}

// FilmGetInput 电影信息查询输入结构
type FilmGetInput struct {
	ID uint `json:"id" invoke-path:"id"`
}

func (f *FilmGetInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if f.ID <= 0 {
		return nerr.FieldError("id", "id can not be lower than 0")
	}

	return nil
}
