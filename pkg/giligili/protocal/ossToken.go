package protocal

import "selfText/giligili_back/libcommon/nerror"

type OSSTokenInput struct {
	Filename string `form:"filename" json:"filename"`
}

func (o *OSSTokenInput) Validate() error {
	nerr := nerror.NewArgumentError("")

	if o.Filename == "" {
		return nerr.FieldError("filename", "filename can not be nil")
	}

	return nil
}
