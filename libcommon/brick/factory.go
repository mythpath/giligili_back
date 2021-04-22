package brick

// Factory create a giligili
type Factory interface {
	// New a giligili instance
	New() interface{}
}

// New return a new object
type NewFunc func() interface{}

// Factory
func FactoryFunc(f NewFunc) Factory {
	return &FactoryFuncWrap{f: f}
}

// FactoryFuncWrap wrap a factory
type FactoryFuncWrap struct {
	f NewFunc
}

func (p *FactoryFuncWrap) New() interface{} {
	return p.f()
}
