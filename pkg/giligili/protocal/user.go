package protocal

import "selfText/giligili_back/libcommon/nerror"

// UserCreateInput 用户创建输入结构
type UserCreateInput struct {
	UserName       string `json:"user_name"`
	PasswordDigest string `json:"password_digest"`
	Nickname       string `json:"nickname"`
	Status         string `json:"status"`
	Avatar         string `json:"avatar"`
}

func (u *UserCreateInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if u.UserName == "" {
		return nerr.FieldError("username", "username cannot be nil")
	}

	return nil
}

// UserUpdateInput 用户更新输入结构
type UserUpdateInput struct {
	ID       uint   `json:"id"`
	UserName string `json:"user_name"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Status   string `json:"status"`
	Avatar   string `json:"avatar"`
}

func (u *UserUpdateInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if u.ID == 0 {
		return nerr.FieldError("username", "id cannot be nil")
	}

	if u.UserName =="" {
		return nerr.FieldError("username","username cannot be nil")
	}

	return nil
}

// UserDeleteInput 用户删除输入结构
type UserDeleteInput struct {
	ID uint `json:"id" invoke-path:"id"`
}

func (u *UserDeleteInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if u.ID== 0 {
		return nerr.FieldError("id", "id cannot be nil")
	}

	return nil
}

// UserGetInput 用户查询输入结构
type UserGetInput struct {
	ID uint `json:"id" invoke-path:"id"`
}

func (u *UserGetInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if u.ID== 0 {
		return nerr.FieldError("id", "id cannot be nil")
	}

	return nil
}

// RegisterInput 注册信息创建输入结构
type RegisterInput struct {
	Nickname        string `form:"nickname" json:"nickname" binding:"required,min=2,max=30"`
	UserName        string `form:"user_name" json:"user_name" binding:"required,min=5,max=30"`
	Password        string `form:"password" json:"password" binding:"required,min=8,max=40"`
	PasswordConfirm string `form:"password_confirm" json:"password_confirm" binding:"required,min=8,max=40"`
}

func (r *RegisterInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if r.Password == "" || r.PasswordConfirm == "" {
		return nerr.FieldError("password", "password or confirmed password can not be nil")
	}

	if r.UserName == "" {
		return nerr.FieldError("userName", "user name can not be nil")
	}

	if r.Nickname == "" {
		return nerr.FieldError("nickName", "nickName cannot be nil")
	}

	return nil
}

// LoginInput 登陆输入信息
type LoginInput struct {
	UserName string `form:"user_name" json:"user_name" binding:"required,min=5,max=30"`
	Password string `form:"password" json:"password" binding:"required,min=8,max=40"`
}

func (l *LoginInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if l.UserName == "" || l.Password == "" {
		return nerr.FieldError("login value", "username and password should both exist")
	}

	return nil
}

// LogoutInput 登出输入信息
type LogoutInput struct {
	ID uint `json:"id" invoke-path:"id"`
}

func (l *LogoutInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if l.ID <= 0 {
		return nerr.FieldError("id", "id can not be lower than 0")
	}

	return nil
}
