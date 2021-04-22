package expr

import (
	"fmt"
)

// Context for expressione
type Context struct {
	data  interface{}
	funcs funcs
}

func (p *Context) call(funcName string, args ...interface{}) (r interface{}, err error) {
	return p.funcs.call(funcName, args...)
}

func (p *Context) value(name string) (interface{}, error) {
	switch d := p.data.(type) {
	case map[string]interface{}:
		return d[name], nil
	}
	return nil, fmt.Errorf("unsupported data")
}
