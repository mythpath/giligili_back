package parallel

import (
	"errors"
	"log"
	"selfText/giligili_back/libcommon/brick"
	"time"
)

// Handle: Callback when the operation of the resource has been completed
type Handle func() error

// FailHandle: ErrorHandle process timeout,
type FailHandle func(err error)

type BackgroundDo struct {
	Config  brick.Config `inject:"config"`
	Timeout time.Duration
}

func (p *BackgroundDo) Init() error {
	p.Timeout = time.Second * time.Duration(p.Config.GetVal("timeout", 10).(float64))
	return nil
}

// BackgroundDo execute f in background.
// t will be called when timeout
func (p *BackgroundDo) BackgroundDo(ok Handle, fail FailHandle) {
	isSuccess := make(chan bool, 1)
	errInfo := make(chan error, 1)
	go func() {
		err := ok()
		if err == nil {
			isSuccess <- true
		} else {
			errInfo <- err
		}
	}()
	go p.doFailMethodIfTimeout(isSuccess, errInfo, fail)
}

func (p *BackgroundDo) doFailMethodIfTimeout(isSuccess chan bool, errInfo chan error, fail FailHandle) {
	select {
	case <-isSuccess:
		log.Println("execute successful!")
	case err := <-errInfo:
		fail(err)
	case <-time.After(p.Timeout):
		fail(errors.New("timeout"))
	}
}
