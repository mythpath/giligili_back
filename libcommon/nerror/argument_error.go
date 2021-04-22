package nerror

import (
	"encoding/json"
)

type ArgumentError struct {
	CommonError
	ArgName     string            `json:"argument"`
	FieldErrors map[string]string `json:"fieldErrors,omitempty"`
}

func NewArgumentError(argName string) *ArgumentError {
	e := &ArgumentError{}
	e.ErrorType = ArgumentErrorType
	e.Msg = "Invalid argument"
	e.ArgName = argName
	e.FieldErrors = make(map[string]string)
	e.StackTrace()
	return e
}

func (p *ArgumentError) FieldError(field, msg string) *ArgumentError {
	p.FieldErrors[field] = extractMsg(msg)
	return p
}

func (p *ArgumentError) Error() string {
	buf, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		buf = []byte(err.Error())
	}
	return string(buf)
}

func (p *ArgumentError) Masking() error {
	e := &ArgumentError{}
	e.ErrorType = p.ErrorType
	e.Msg = p.Msg
	e.ArgName = p.ArgName
	e.FieldErrors = p.FieldErrors
	return e
}

func (p *ArgumentError) Type() string {
	return p.ErrorType
}
