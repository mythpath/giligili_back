package fsm

import (
	"context"
	"fmt"
	"log"
)

// Group is set of states and transitions
type Group interface {
	// WatchEvents receives the event when the state is being changed.
	WatchEvents() chan Event
	// Submit submit a transtion
	Submit(ctx context.Context, action string, arg interface{}) error

	// GetID return the id of group
	GetID() uint
}

// NewGroup returns an instance of state group
func NewGroup(states map[string]StateFrame, transition []Transition, redoLogger RedoLogger, id uint, presentState StateFrame) (Group, error) {

	//validate

	if states == nil || len(states) == 0 {
		return nil, fmt.Errorf("states can not be nil")
	}
	if transition == nil || len(transition) == 0 {
		return nil, fmt.Errorf("transition can not be nil")
	}
	if redoLogger == nil {
		return nil, fmt.Errorf("RedoLogger can not be nil")
	}

	eventChan := make(chan Event, 1)
	group := &group{
		id:           id,
		states:       states,
		trans:        transition,
		events:       eventChan,
		redoLogger:   redoLogger,
		presentState: presentState,
		lockChan:     make(chan int, 1),
	}
	go group.Start()
	return group, nil
}

type group struct {
	id           uint //id of group
	states       map[string]StateFrame
	trans        []Transition
	events       chan Event
	presentState StateFrame
	redoLogger   RedoLogger
	lockChan     chan int // lock group until completed or failed
}

//GetTrsnsitions return all transtions in current state
func (p *group) GetTransitions(from State) []Transition {
	var trasition []Transition
	for _, tran := range p.trans {
		if tran.From.State == from {
			trasition = append(trasition, tran)
		}
	}
	return trasition
}

// Submit a transtion
func (p *group) Submit(ctx context.Context, action string, arg interface{}) error {
	p.lockChan <- 1
	transition := p.getTransitionByAction(action)

	if transition.IsEmpty() {
		<-p.lockChan
		return fmt.Errorf("error action [%s] for fsm present state [%d]", action, p.presentState.State)
	}

	event, err := p.redoLogger.Commit(transition.From, transition.To, transition.Fail, arg, action, Doing, p.id)
	if err != nil {
		<-p.lockChan
		return err
	}

	p.raiseEvent(event)
	return nil
}

func (p *group) getTransitionByAction(action string) Transition {
	var transition Transition
	for _, tran := range p.trans {
		if tran.Action == action && tran.From.State == p.presentState.State {
			transition = tran
			break
		}
	}
	return transition
}

// Start group
func (p *group) Start() {
	//recovery
	// for {
	// 	redoLog, err := p.redoLogger.Redo()
	// 	if err != nil {
	// 		log.Print(err)
	// 	}
	// 	select {
	// 	case redoEvent := <-redoLog:
	// 		log.Print("start of redo is : ", redoEvent.From(), redoEvent.To())
	// 		p.raiseEvent(redoEvent)
	// 	default:
	// 	}
	// }

}

func (p *group) WatchEvents() chan Event {
	return p.events
}

func (p *group) raiseEvent(event Event) {
	p.presentState = event.From()
	event.SetGroup(p)
	p.events <- event
}

func (p *group) complete(e Event, data interface{}) {
	event, err := p.redoLogger.Commit(e.To(), StateFrame{}, StateFrame{}, data, e.Action(), Done, p.id)
	if err != nil {
		log.Print(err)
		<-p.lockChan
		return
	}
	p.raiseEvent(event)
	p.redoLogger.Apply(event.GetEventID())
	<-p.lockChan
}

func (p *group) fail(e Event, data interface{}) {
	event, err := p.redoLogger.Commit(e.GetFail(), StateFrame{}, StateFrame{}, data, e.Action(), Done, p.id)
	if err != nil {
		log.Print(err)
		<-p.lockChan
		return
	}
	p.raiseEvent(event)
	p.redoLogger.Apply(event.GetEventID())
	<-p.lockChan
}

func (p *group) GetID() uint {
	return p.id
}
