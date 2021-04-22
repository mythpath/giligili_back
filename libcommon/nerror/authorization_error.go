package nerror

import (
	"encoding/json"
	"fmt"
)

type AuthorizationError struct {
	CommonError
}

func NewAuthorizationError(msgFormat string, a ...interface{}) *AuthorizationError {
	e := &AuthorizationError{}
	e.ErrorType = AuthorizationErrorType
	e.Msg = fmt.Sprintf(msgFormat, a...)
	e.StackTrace()
	return e
}

func NewAuthorizationErrorForward(data []byte) *AuthorizationError {
	e := &AuthorizationError{}
	e.ErrorType = AuthorizationErrorType
	if json.Valid(data) {
		json.Unmarshal(data, e)
	}
	e.StackTrace()
	return e
}

func (p *AuthorizationError) Error() string {
	return p.CommonError.Error()
}

func (p *AuthorizationError) Masking() error {
	return p.CommonError.Masking()
}

func (p *AuthorizationError) Type() string {
	return p.ErrorType
}
