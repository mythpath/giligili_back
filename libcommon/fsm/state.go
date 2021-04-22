package fsm

// State of a group. User must define concrete state from the type
type StateFrame struct {
	State State
	Final bool // Final is final state or not
}

type State int

const Init State = 0
