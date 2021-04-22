package fsm_test

import (
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/fsm"
	"selfText/giligili_back/libcommon/orm/dialects/mysql"
	"flag"
	"log"
	"math/rand"
	"testing"
	"time"
)

const (
	stateInit = fsm.Init
	state1    = fsm.Init + 1
	state2    = fsm.Init + 2
	state3    = fsm.Init + 3
	state4    = fsm.Init + 4
)

const (
	stateInitName = "stateInit"
	state1Name    = "state1"
	state2Name    = "state2"
	state3Name    = "state3"
	state4Name    = "state4"
)

const (
	action1    = "action1"
	action2    = "action2"
	action3    = "action3"
	actionQuit = "actionQuit"
)

func doAction1() bool {
	i := rand.Intn(2)
	if i == 0 {
		log.Printf("doAction1 failed")
		return false
	}
	log.Printf("doAction1 successed")
	return true
}

func doAction2() bool {
	i := rand.Intn(2)
	if i == 0 {
		log.Printf("doAction2 failed")
		return false
	}
	log.Printf("doAction2 successed")
	return true
}

func doAction3() bool {
	i := rand.Intn(2)
	if i == 0 {
		log.Printf("doAction3 failed")
		return false
	}
	log.Printf("doAction3 successed")
	return true
}

func doActionQuit() bool {
	return true
}

func TestConcurrency(t *testing.T) {
	configPath := flag.String("c", "config.json", "configuration file")

	container := brick.NewContainer()
	container.Add(&brick.JSONConfigService{}, "config", brick.FactoryFunc(func() interface{} {
		return brick.NewJSONConfigService(*configPath)
	}))

	container.Add(&mysql.MySQLService{}, "DB", nil)
	container.Add(&fsm.RedoPersistService{}, "redo", nil)
	container.Add(&StateMachine{}, "StateMachine", nil)
	container.Build()

	c := container.GetByName("StateMachine").(*StateMachine)

	c.Signal(action1)
	c.Signal(action1)
	c.Signal(action1)
	c.Signal(action2)
	c.Signal(action2)
	c.Signal(action2)
	c.Signal(action3)
	c.Signal(action3)
	c.Signal(action3)

	select {}
}

type StateMachine struct {
	Redo  *fsm.RedoPersistService `inject:"redo"`
	group fsm.Group
}

func (p *StateMachine) Init() error {
	state := map[string]fsm.StateFrame{
		stateInitName: fsm.StateFrame{State: stateInit, Final: false},
		state1Name:    fsm.StateFrame{State: state1, Final: false},
		state2Name:    fsm.StateFrame{State: state2, Final: false},
		state3Name:    fsm.StateFrame{State: state3, Final: false},
	}
	var trasitions []fsm.Transition

	t1Trasition := fsm.Transition{
		From:   state[stateInitName],
		To:     state[state1Name],
		Fail:   state[state2Name],
		Action: action1,
	}

	t2Trasition := fsm.Transition{
		From:   state[stateInitName],
		To:     state[state2Name],
		Fail:   state[state3Name],
		Action: action2,
	}

	t3Trasition := fsm.Transition{
		From:   state[state3Name],
		To:     state[state2Name],
		Fail:   state[state1Name],
		Action: action3,
	}

	t4Trasition := fsm.Transition{
		From:   state[state2Name],
		To:     state[state1Name],
		Fail:   state[state3Name],
		Action: action2,
	}

	t5Trasition := fsm.Transition{
		From:   state[state1Name],
		To:     state[state2Name],
		Fail:   state[state1Name],
		Action: action1,
	}

	trasitions = append(trasitions, t1Trasition)
	trasitions = append(trasitions, t2Trasition)
	trasitions = append(trasitions, t3Trasition)
	trasitions = append(trasitions, t4Trasition)
	trasitions = append(trasitions, t5Trasition)
	rand.Seed(time.Now().UTC().UnixNano())
	group, err := fsm.NewGroup(state, trasitions, p.Redo, uint(rand.Uint32()), fsm.StateFrame{
		State: stateInit,
		Final: false,
	})
	if err != nil {
		log.Print(err)
		return err
	}
	p.group = group
	log.Printf("start from %s state", stateInitName)

	go p.loop()

	return nil
}

func (p *StateMachine) Signal(action string) {
	go func(action string) {
		log.Printf("send %s signal", action)
		err := p.group.Submit(nil, action, nil)
		if err != nil {
			log.Printf("signal failed: %s", err.Error())
		}
	}(action)
}

func (p *StateMachine) loop() {
	for {
		select {
		case event := <-p.group.WatchEvents():
			p.eventSwitch(event)
		}
	}
}

func (p *StateMachine) eventSwitch(event fsm.Event) {
	switch event.From().State {
	case stateInit:
		if event.Status() == fsm.Doing {
			log.Printf("%s leaving with action: %s", stateInitName, event.Action())
			if event.Action() == action1 {
				if doAction1() {
					event.Complete(nil)
					log.Printf("should go to %d", event.To().State)
				} else {
					event.Fail(nil)
					log.Printf("should go to %d", event.GetFail().State)
				}
			} else if event.Action() == action2 {
				if doAction2() {
					event.Complete(nil)
					log.Printf("should go to %d", event.To().State)
				} else {
					event.Fail(nil)
					log.Printf("should go to %d", event.GetFail().State)
				}
			}
		} else {
			log.Println("=== === === ===")
			log.Printf("%s enter with action: %s", stateInitName, event.Action())
		}
	case state1:
		if event.Status() == fsm.Doing {
			log.Printf("%s leaving with action: %s", state1Name, event.Action())
			if doAction1() {
				event.Complete(nil)
				log.Printf("should go to %d", event.To().State)
			} else {
				event.Fail(nil)
				log.Printf("should go to %d", event.GetFail().State)
			}
		} else {
			log.Println("=== === === ===")
			log.Printf("%s enter with action: %s", state1Name, event.Action())
		}
	case state2:
		if event.Status() == fsm.Doing {
			log.Printf("%s leaving with action: %s", state2Name, event.Action())
			if doAction2() {
				event.Complete(nil)
				log.Printf("should go to %d", event.To().State)
			} else {
				event.Fail(nil)
				log.Printf("should go to %d", event.GetFail().State)
			}
		} else {
			log.Println("=== === === ===")
			log.Printf("%s enter with action: %s", state2Name, event.Action())
		}
	case state3:
		if event.Status() == fsm.Doing {
			log.Printf("%s leaving with action %s", state3Name, event.Action())
			if doAction3() {
				event.Complete(nil)
				log.Printf("should go to %d", event.To().State)
			} else {
				event.Fail(nil)
				log.Printf("should go to %d", event.GetFail().State)
			}
		} else {
			log.Println("=== === === ===")
			log.Printf("%s enter with action: %s", state3Name, event.Action())
			p.Signal(action3)
		}
	}
}
