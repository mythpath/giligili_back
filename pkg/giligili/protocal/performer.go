package protocal

import (
	"selfText/giligili_back/libcommon/nerror"
)

// PerformerCreateInput 演员信息创建输入结构
type PerformerCreateInput struct {
	Name    string `json:"name"`
	Chinese string `json:"chinese"`
	Gender  int    `binding:"required min=0,max=1" json:"gender"`
	Locate  string `json:"locate"`
	Avatar  string `gorm:"size:1000" json:"avatar"`
	Rank    string `json:"rank"`
	Comment string `json:"comment"`
}

func (p *PerformerCreateInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if p.Gender < 0 || p.Gender > 1 {
		return nerr.FieldError("gender", "gender can only be 0 or 1")
	}

	if p.Name == "" {
		return nerr.FieldError("name", "name can not be nil")
	}

	return nil
}

// PerformerDeleteInput 演员信息删除输入结构
type PerformerDeleteInput struct {
	ID uint `json:"id" invoke-path:"id"`
}

func (p *PerformerDeleteInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if p.ID <= 0 {
		return nerr.FieldError("id", "id can not be lower than 0")
	}

	return nil
}

// PerformerUpdateInput 演员信息更新输入结构
type PerformerUpdateInput struct {
	ID uint `json:"id" invoke-path:"id"`
	PerformerCreateInput
}

func (p *PerformerUpdateInput) Validate() error {
	if err := p.PerformerCreateInput.Validate(); err != nil {
		return err
	}
	nerr := nerror.NewArgumentError("")

	if p.Gender < 0 || p.Gender > 1 {
		return nerr.FieldError("gender", "gender can only be 0 or 1")
	}

	if p.Name == "" {
		return nerr.FieldError("name", "name can not be nil")
	}

	return nil
}

// PerformerGetInput 演员信息查询输入结构
type PerformerGetInput struct {
	ID uint `json:"id" invoke-path:"id"`
}

func (p *PerformerGetInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if p.ID <= 0 {
		return nerr.FieldError("id", "id can not be lower than 0")
	}

	return nil
}
