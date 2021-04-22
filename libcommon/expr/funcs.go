package expr

import (
	"fmt"
	"reflect"
	"unicode"
)

var (
	errorType = reflect.TypeOf((*error)(nil)).Elem()
)

type funcs map[string]reflect.Value

func newFuncs(fns map[string]interface{}) funcs {
	out := funcs{}
	for name, fn := range fns {
		if !goodName(name) {
			panic(fmt.Errorf("function name %s is not a valid identifier", name))
		}
		v := reflect.ValueOf(fn)
		if v.Kind() != reflect.Func {
			panic("value for " + name + " not a function")
		}
		if !goodFunc(v.Type()) {
			panic(fmt.Errorf("can't install method/function %q with %d results", name, v.Type().NumOut()))
		}
		out[name] = v
	}
	return out
}

func goodName(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		switch {
		case r == '_':
		case i == 0 && !unicode.IsLetter(r):
			return false
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			return false
		}
	}
	return true
}

func goodFunc(typ reflect.Type) bool {
	switch {
	case typ.NumOut() == 2 && typ.Out(1) == errorType:
		return true
	}
	return false
}

func (p *funcs) call(name string, args ...interface{}) (interface{}, error) {
	var in []reflect.Value
	if fn, ok := p.findFunction(name); ok {
		for _, arg := range args {
			in = append(in, reflect.ValueOf(arg))
		}
		rets := fn.Call(in)
		v2 := rets[1].Interface()
		if v2 != nil {
			err, ok := v2.(error)
			if !ok {
				return nil, fmt.Errorf("the second must be error")
			}
			if err != nil {
				return nil, err
			}
		}
		
		return rets[0].Interface(), nil
	}
	return nil, fmt.Errorf("could not found function %s", name)
}

func (p *funcs) findFunction(name string) (reflect.Value, bool) {
	if fn := (*p)[name]; fn.IsValid() {
		return fn, true
	}
	return reflect.Value{}, false
}
