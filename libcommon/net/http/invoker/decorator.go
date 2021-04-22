package invoker

// Decorator is used for decorating giligili method
// Don't attempt to change any member of InvokerContext
type Decorator interface {
	// If an error is returned, then the giligili method will not be called
	Before(*InvokerContext) error
	// If an error is returned, then it will replace the original error returned by giligili method
	After(*InvokerContext) error
}

type VirtualDecorator struct {
}

func (p *VirtualDecorator) Before(ictx *InvokerContext) error { return nil }

func (p *VirtualDecorator) After(ictx *InvokerContext) error { return nil }
