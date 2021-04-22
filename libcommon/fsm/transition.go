package fsm

// Transition defines the path of the conversion between two state
type Transition struct {
	// From the state where the action can't be done
	From StateFrame
	// To the state if complete
	To StateFrame
	// Fail the state if failed
	Fail StateFrame
	// Action do the transition
	Action string
}

func (p Transition) IsEmpty() bool {

	if p.From.State == 0 && p.To.State == 0 && p.Fail.State == 0 && p.Action == "" {
		return true
	}
	return false
}
