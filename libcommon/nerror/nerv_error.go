package nerror

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
)

const (
	CommonErrorType        = "StandardError"
	PanicErrorType         = "PanicError"
	ArgumentErrorType      = "ArgumentError"
	AuthorizationErrorType = "AuthorizationError"
)

type NervError interface {
	Masking() error
	Type() string
}

type frame struct {
	Func string
	File string
	Line int
}

type CommonError struct {
	ErrorType string   `json:"errType,omitempty"`
	Msg       string   `json:"msg,omitempty"`
	Stack     []*frame `json:"stack,omitempty"`
	Detail    string   `json:"-"`
}

func NewCommonError(msgFormat string, a ...interface{}) *CommonError {
	e := &CommonError{}
	e.ErrorType = CommonErrorType
	var nerr *CommonError
	for i, _ := range a {
		switch a[i].(type) {
		case *CommonError:
			nerr = a[i].(*CommonError)
			a[i] = nerr.Msg
		case string:
			errStr := a[i].(string)
			a[i] = extractMsg(errStr)
		}
	}
	if strings.Contains(msgFormat, "errType") {
		msgFormat = extractMsg(msgFormat)
	}
	e.Msg = fmt.Sprintf(msgFormat, a...)
	if nerr != nil {
		e.Stack = nerr.Stack
	} else {
		e.StackTrace()
	}
	return e
}

func WrapCommonError(err error) *CommonError {
	e := &CommonError{
		ErrorType: CommonErrorType,
		Msg:       err.Error(),
	}

	return e
}

func (p *CommonError) Error() string {
	stack, _ := json.MarshalIndent(p.Stack, "", "  ")
	str := "\n" + "errType: " + p.ErrorType + "\n"
	str += "msg: " + p.Msg + "\n"
	str += "detail: " + p.Detail + "\n"
	str += "stack: " + string(stack)
	return str
}

func (p *CommonError) Masking() error {
	e := &CommonError{}
	e.ErrorType = p.ErrorType
	e.Msg = p.Msg
	return e
}

func (p *CommonError) Type() string {
	return p.ErrorType
}

func (p *CommonError) StackTrace() *CommonError {
	for i := 2; ; i++ {
		funcName, file, line, ok := runtime.Caller(i)
		if ok {
			function := runtime.FuncForPC(funcName).Name()
			frame := &frame{
				Func: function,
				File: file,
				Line: line,
			}
			p.Stack = append(p.Stack, frame)
		} else {
			break
		}
	}
	return p
}

func extractMsg(errStr string) string {
	msg := errStr
	if strings.Contains(errStr, "errType") {
		errReader := bufio.NewReader(bytes.NewBufferString(errStr))
		for {
			line, _, err := errReader.ReadLine()
			if err != nil {
				break
			}
			if strings.HasPrefix(string(line), "msg") {
				msg = string(line)
				break
			}
		}
	}
	return msg
}
