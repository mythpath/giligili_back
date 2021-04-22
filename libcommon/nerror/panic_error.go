package nerror

import (
	"encoding/json"
	"fmt"
)

type PanicError struct {
	CommonError
	Panic interface{} `json:"panic,omitempty"`
}

func NewPanicError(r interface{}) *PanicError {
	e := &PanicError{}
	e.Stack = make([]*frame, 0)
	e.ErrorType = PanicErrorType
	e.Msg = "Server met some problem"
	e.Panic = r
	e.StackTrace()
	return e
}

func (p *PanicError) Error() string {
	stack, _ := json.MarshalIndent(p.Stack, "", "  ")
	str := "\n" + "errType: " + p.ErrorType + "\n"
	str += "msg: " + p.Msg + "\n"
	str += "panic: " + fmt.Sprintf("%v\n", p.Panic)
	str += "stack: " + string(stack)
	return str
}

func (p *PanicError) Masking() error {
	return p.CommonError.Masking()
}

func (p *PanicError) Type() string {
	return p.ErrorType
}
