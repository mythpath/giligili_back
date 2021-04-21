package protocal

import "selfText/giligili_back/libcommon/nerror"

// VideoCreateInput 视频信息创建输入结构
type VideoCreateInput struct {
	Title  string `json:"title"`
	Info   string `json:"info"`
	Url    string `json:"url"`
	Avatar string `json:"avatar"`
	UserID uint   `json:"user_id"`
}

func (v *VideoCreateInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if v.Title == "" {
		return nerr.FieldError("title", "title con not be nil")
	}

	if v.Url == "" {
		return nerr.FieldError("url", "url con not be nil")
	}

	if v.UserID == 0 {
		return nerr.FieldError("userID", "userID con not be nil")
	}

	return nil
}

// VideoDeleteInput 视频信息删除输入结构
type VideoDeleteInput struct {
	ID uint `json:"id" invoke-path:"id"`
}

func (v *VideoDeleteInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if v.ID <= 0 {
		return nerr.FieldError("id", "id can not be lower than 0")
	}

	return nil
}

// VideoUpdateInput 视频信息更新输入结构
type VideoUpdateInput struct {
	ID uint `json:"id" invoke-path:"id"`
	VideoCreateInput
}

func (v *VideoUpdateInput) Validate() error {
	if err := v.VideoCreateInput.Validate(); err != nil {
		return err
	}
	nerr := nerror.NewArgumentError("")

	if v.ID <= 0 {
		return nerr.FieldError("id", "id can not be lower than 0")
	}

	return nil
}

// VideoGetInput 视频信息查询输入结构
type VideoGetInput struct {
	ID uint `json:"id" invoke-path:"id"`
}

func (v *VideoGetInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if v.ID <= 0 {
		return nerr.FieldError("id", "id can not be lower than 0")
	}

	return nil
}
