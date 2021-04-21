package protocal

import "selfText/giligili_back/libcommon/nerror"

// TimeCreateInput 时间服务器记录输入结构
type TimeCreateInput struct {
	URL     string `json:"url"`
	Comment string  `json:"comment"`
}

func (t *TimeCreateInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if t.URL == ""  {
		return nerr.FieldError("url", "time server url should not be null.")
	}

	return nil
}

// TimeDeleteInput 时间服务器记录删除输入结构
type TimeDeleteInput struct {
	ID uint `json:"id" invoke-path:"id"`
}

func (t *TimeDeleteInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if t.ID == 0  {
		return nerr.FieldError("id", "record id of the time server should not be zero.")
	}

	return nil
}

// TimeGetInput 时间服务器记录获取输入结构
type TimeGetInput struct {
	ID uint `json:"id" invoke-path:"id"`
}

func (t *TimeGetInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if t.ID == 0  {
		return nerr.FieldError("id", "record id of the time server should not be zero.")
	}

	return nil
}

// TimeNowInput 时间服务器记录获取输入结构
type TimeNowInput struct {
	ID uint `json:"id" invoke-path:"id"`
}

func (t *TimeNowInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if t.ID == 0  {
		return nerr.FieldError("id", "record id of the time server should not be zero.")
	}

	return nil
}
