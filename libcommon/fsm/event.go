package fsm

// Event notify a state transition
type Event interface {
	//From is source state
	From() StateFrame

	//To is target state
	To() StateFrame

	Data() interface{}

	//Action return the action of event
	Action() string

	// Complete the transition after the action has been done
	Complete(data interface{})
	Status() Status
	// GetFail return the fail state
	GetFail() StateFrame

	// Fail the transition when the action can't be completed or timeout
	Fail(data interface{})
	// GetEventID return id of event of fsm
	GetEventID() uint

	//SetGroup set a gourp for event
	SetGroup(group *group)
}

type Status int

const Done Status = 0
const Doing Status = 1

type event struct {
	id     uint
	group  *group
	from   StateFrame
	to     StateFrame
	fail   StateFrame
	data   interface{}
	action string
	status Status //doing|done
}

func (p *event) From() StateFrame {
	return p.from
}

func (p *event) To() StateFrame {
	return p.to
}

func (p *event) Data() interface{} {
	return p.data
}

func (p *event) GetFail() StateFrame {
	return p.fail
}

func (p *event) Status() Status {
	return p.status
}

func (p *event) Action() string {
	return p.action
}

func (p *event) Complete(data interface{}) {

	p.group.redoLogger.Apply(p.id)
	p.group.complete(p, data)

}

func (p *event) Fail(data interface{}) {
	p.group.redoLogger.Apply(p.id)
	p.group.fail(p, data)

}

func (p *event) GetEventID() uint {
	return p.id
}

func (p *event) SetGroup(group *group) {
	p.group = group
}
