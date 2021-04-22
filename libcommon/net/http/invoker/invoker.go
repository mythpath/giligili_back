package invoker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"selfText/giligili_back/libcommon/nerror"
	"selfText/giligili_back/libcommon/orm"
	"selfText/giligili_back/libcommon/passport/auth"
	"strconv"

	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/net/http/rest"
	"selfText/giligili_back/libcommon/net/http/rest/render"

	"github.com/go-chi/chi"
)

type Invoker struct {
	container *brick.Container
	Passport  *auth.PassportClient `inject:"passport-client"`

	decorator        Decorator
	decorateEnable   map[uintptr]bool
	decorateOnMethod map[uintptr]Decorator
	internalErr      chan error
}

func NewInvoker(d Decorator, exErr chan error) *Invoker {
	p := &Invoker{}
	p.decorator = d
	p.decorateEnable = make(map[uintptr]bool)
	p.decorateOnMethod = make(map[uintptr]Decorator)
	if exErr != nil {
		p.internalErr = exErr
	} else {
		p.internalErr = make(chan error)
	}

	return p
}

// SetContainer call by container
func (p *Invoker) SetContainer(c *brick.Container) {
	p.container = c
}

// SetDecorator set default decorator for all registered giligili method,
// unless DisableDecoratorOn is called
func (p *Invoker) SetDecorator(d Decorator) {
	p.decorator = d
}

func (p *Invoker) Init() error {
	if p.decorateEnable == nil {
		p.decorateEnable = make(map[uintptr]bool)
	}
	if p.decorateOnMethod == nil {
		p.decorateOnMethod = make(map[uintptr]Decorator)
	}

	return nil
}

type InvokerContext struct {
	container    *brick.Container
	req          *http.Request
	w            http.ResponseWriter
	httpParsers  []ParserHandler
	fn           *reflect.Value
	funcParams   []reflect.Value
	funcReturns  []reflect.Value
	needAuth     bool
	moduleName   string
	resourceName string
	decorator    Decorator
}

// Execute an operation by call the method of the class that has been registered in the container
// To parse application/json, should use tag `json`
// To parse URL query string, should use tag `invoke-query`
// To parse URL path parameter, should use tag `invoke-path`
// To parse HTTP header, should use tag `invoke-header`
//
// check invoker_test.go for detail usage
func (p *Invoker) Execute(class, method string) http.HandlerFunc {
	fn, params := p.panicMethod(p.container, class, method)
	panicMethodOut(fn, class, method)
	parsers := []ParserHandler{
		parseInvokeQuery, parseInvokePath, parseInvokejson, parseInvokeHeader,
	}
	return func(w http.ResponseWriter, req *http.Request) {
		defer p.handleInvokerPanic(w, req)
		ictx := &InvokerContext{
			container:   p.container,
			req:         req,
			w:           w,
			httpParsers: parsers,
			fn:          &fn,
			funcParams:  make([]reflect.Value, len(params)),
			funcReturns: make([]reflect.Value, 2),
		}
		copy(ictx.funcParams, params)
		httpBody, httpCode := p.invokeMethod(ictx)

		render.Status(req, httpCode)
		render.JSON(w, req, httpBody)
	}
}

func (p *Invoker) ExecuteWithAuth(class, method, moduleName, rscName string) http.HandlerFunc {
	if p.Passport == nil {
		panic("No passport destination connection")
	}
	fn, params := p.panicMethod(p.container, class, method)
	panicMethodOut(fn, class, method)
	return func(w http.ResponseWriter, req *http.Request) {
		defer p.handleInvokerPanic(w, req)
		ictx := &InvokerContext{
			container: p.container,
			req:       req,
			w:         w,
			httpParsers: []ParserHandler{
				parseInvokeQuery, parseInvokePath, parseInvokejson, parseInvokeHeader,
			},
			fn:           &fn,
			funcParams:   make([]reflect.Value, len(params)),
			funcReturns:  make([]reflect.Value, 2),
			needAuth:     true,
			moduleName:   moduleName,
			resourceName: rscName,
		}
		copy(ictx.funcParams, params)
		httpBody, httpCode := p.invokeMethod(ictx)
		render.Status(req, httpCode)
		render.JSON(w, req, httpBody)
	}
}

func (p *Invoker) ExecuteWithAuthAndStatus(class, method, moduleName, rscName string, status int) http.HandlerFunc {
	if p.Passport == nil {
		panic("No passport destination connection")
	}
	fn, params := p.panicMethod(p.container, class, method)
	panicMethodOut(fn, class, method)
	return func(w http.ResponseWriter, req *http.Request) {
		defer p.handleInvokerPanic(w, req)
		ictx := &InvokerContext{
			container: p.container,
			req:       req,
			w:         w,
			httpParsers: []ParserHandler{
				parseInvokeQuery, parseInvokePath, parseInvokejson, parseInvokeHeader,
			},
			fn:           &fn,
			funcParams:   make([]reflect.Value, len(params)),
			funcReturns:  make([]reflect.Value, 2),
			needAuth:     true,
			moduleName:   moduleName,
			resourceName: rscName,
		}
		copy(ictx.funcParams, params)
		httpBody, httpCode := p.invokeMethodWithStatus(ictx, status)
		render.Status(req, httpCode)
		render.JSON(w, req, httpBody)
	}
}

func (p *Invoker) ExecuteText(class, method string) http.HandlerFunc {
	fn, params := p.panicMethod(p.container, class, method)
	panicTextMethodOut(fn, class, method)
	return func(w http.ResponseWriter, req *http.Request) {
		defer p.handleInvokerPanic(w, req)
		ictx := &InvokerContext{
			container: p.container,
			req:       req,
			w:         w,
			httpParsers: []ParserHandler{
				parseInvokeQuery, parseInvokePath, parseInvokejson, parseInvokeHeader,
			},
			fn:          &fn,
			funcParams:  make([]reflect.Value, len(params)),
			funcReturns: make([]reflect.Value, 2),
		}
		copy(ictx.funcParams, params)
		httpBody, httpCode := p.invokeMethod(ictx)
		if httpCode != 200 {
			render.Status(req, httpCode)
			render.JSON(w, req, httpBody)
		} else {
			w.WriteHeader(200)
			w.Header().Set("Content-type", "text/plain")
			w.Write(httpBody.([]byte))
		}
	}
}

func (p *Invoker) ExecuteTextWithAuth(class, method, moduleName, rscName string) http.HandlerFunc {
	if p.Passport == nil {
		panic("No passport destination connection")
	}
	fn, params := p.panicMethod(p.container, class, method)
	panicTextMethodOut(fn, class, method)
	return func(w http.ResponseWriter, req *http.Request) {
		defer p.handleInvokerPanic(w, req)
		ictx := &InvokerContext{
			container: p.container,
			req:       req,
			w:         w,
			httpParsers: []ParserHandler{
				parseInvokeQuery, parseInvokePath, parseInvokejson, parseInvokeHeader,
			},
			fn:           &fn,
			funcParams:   make([]reflect.Value, len(params)),
			funcReturns:  make([]reflect.Value, 2),
			needAuth:     true,
			moduleName:   moduleName,
			resourceName: rscName,
		}
		copy(ictx.funcParams, params)
		httpBody, httpCode := p.invokeMethod(ictx)
		if httpCode != 200 {
			render.Status(req, httpCode)
			render.JSON(w, req, httpBody)
		} else {
			w.WriteHeader(200)
			w.Header().Set("Content-type", "text/plain")
			w.Write(httpBody.([]byte))
		}
	}
}

// SetDecoratorOn should be called after Executexxx
func (p *Invoker) SetDecoratorOn(class, method string, d Decorator) {
	fn, _, err := invokeGetMethod(p.container, class, method)
	if err != nil {
		panic(err.Error())
	}

	// It's safe to use Pointer() to indentify an unique function in this circumstance,
	// becuase the function is preloaded.
	p.decorateEnable[fn.Pointer()] = true
	p.decorateOnMethod[fn.Pointer()] = d
}

// DisableDecoratorOn should be called after Executexxx
func (p *Invoker) DisableDecoratorOn(class, method string) {
	fn, _, err := invokeGetMethod(p.container, class, method)
	if err != nil {
		panic(err.Error())
	}

	p.decorateEnable[fn.Pointer()] = false
}

func (p *Invoker) panicMethod(c *brick.Container, class, method string) (reflect.Value, []reflect.Value) {
	fn, params, err := invokeGetMethod(c, class, method)
	if err != nil {
		panic(err.Error())
	}

	// ignore duplicate
	p.decorateEnable[fn.Pointer()] = true
	panicMethodIn(fn, class, method)
	return fn, params
}

func panicMethodIn(fn reflect.Value, class, method string) {
	if fn.Type().NumIn() < 2 {
		s := fmt.Sprintf("Unsupported function[%s.%s], parameter is empty, "+
			"should have pattern like (ctx context.Context, arg StructA)",
			class, method)
		panic(s)
	}
	argType := fn.Type().In(1)
	if argType.String() != "context.Context" {
		s := fmt.Sprintf("Unsupported function[%s.%s] parameter[%d]: %s, "+
			"should have pattern like (ctx context.Context, arg StructA)",
			class, method, 1, argType.String())
		panic(s)
	}
	for paramIdx := 2; paramIdx < fn.Type().NumIn(); paramIdx++ {
		argType := fn.Type().In(paramIdx)
		if argType.Kind() == reflect.Struct ||
			argType.Kind() == reflect.Slice ||
			(argType.Kind() == reflect.Ptr && argType.Elem().Kind() == reflect.Struct) {
			continue
		} else {
			s := fmt.Sprintf("Unsupported function[%s.%s] parameter[%d]: %s, "+
				"should have pattern like (ctx context.Context, arg StructA), "+
				"or like (ctx context.Context, args []StructA)",
				class, method, paramIdx, argType.String())
			panic(s)
		}
	}
}

func panicMethodOut(fn reflect.Value, class, method string) {
	if fn.Type().NumOut() != 2 || fn.Type().Out(1).String() != "error" {
		errPattern := fmt.Sprintf("current function: %s.%s, return:", class, method)
		for i := 0; i < fn.Type().NumOut(); i++ {
			errPattern = fmt.Sprintf("%s %s, ", errPattern, fn.Type().Out(i).String())
		}
		s := fmt.Sprintf("invoked funciton should have only 2 return values, "+
			"while the second should be of type error: %s",
			errPattern)
		panic(s)
	}
}

func panicTextMethodOut(fn reflect.Value, class, method string) {
	if fn.Type().NumOut() != 2 ||
		fn.Type().Out(0).String() != "[]uint8" ||
		fn.Type().Out(1).String() != "error" {
		errPattern := fmt.Sprintf("current function: %s.%s, return:", class, method)
		for i := 0; i < fn.Type().NumOut(); i++ {
			errPattern = fmt.Sprintf("%s %s, ", errPattern, fn.Type().Out(i).String())
		}
		s := fmt.Sprintf("invoked funciton should have only 2 return values, "+
			"while the first shoud be of type []uint8 or []byte, "+
			"and the second should be of type error: %s",
			errPattern)
		panic(s)
	}
}

// ParserHandler parses informations from http request and put them in "argp"
type ParserHandler func(req *http.Request, argType reflect.Type, argp *reflect.Value) error

func (p *Invoker) handleInvokerPanic(w http.ResponseWriter, req *http.Request) {
	if r := recover(); r != nil {
		nerr := nerror.NewPanicError(r)
		if p.internalErr != nil {
			select {
			case p.internalErr <- nerr:
			default:
				fmt.Printf("panic: %v\n", nerr.Error())
			}
		} else {
			fmt.Printf("panic: %v\n", nerr.Error())
		}
		render.Status(req, 500)
		render.JSON(w, req, nerr.Masking())
	}
}

func (p *Invoker) sendError(err error) {
	if p.internalErr != nil {
		select {
		case p.internalErr <- err:
		default:
		}
	}
}

func (p *Invoker) invokeMethod(ictx *InvokerContext) (body interface{}, code int) {
	validate := func(argp *reflect.Value, body *interface{}, code *int) {
		if v, ok := argp.Interface().(Validator); ok {
			err := v.Validate()
			if err != nil {
				p.sendError(err)
				*code = http.StatusBadRequest
				if ne, ok := err.(nerror.NervError); ok {
					*body = ne.Masking()
				} else {
					*body = err
				}
				return
			}
		}
		return
	}
	withAuth := func(body *interface{}, code *int) *reflect.Value {
		user, clientToken := readHeader(ictx.req)
		info, err := p.Passport.CheckAuthorization(user, "", ictx.moduleName, ictx.resourceName,
			clientToken, p.Passport.GetModuleToken(ictx.moduleName), 0)
		if err != nil {
			p.sendError(err)
			*code = http.StatusBadRequest
			if ne, ok := err.(nerror.NervError); ok {
				resolveError(ne, body, code)
			} else {
				*body = err
			}
			return nil
		}
		value := setContext(user, info)

		return value
	}
	withoutAuth := func(body *interface{}, code *int) *reflect.Value {
		user, _ := readHeader(ictx.req)
		value := setContext(user, nil)

		return value
	}

	for paramIdx := 1; paramIdx < ictx.fn.Type().NumIn(); paramIdx++ {
		argType := ictx.fn.Type().In(paramIdx)
		if argType.String() == "context.Context" {
			var argp *reflect.Value
			if ictx.needAuth {
				argp = withAuth(&body, &code)
			} else {
				argp = withoutAuth(&body, &code)
			}
			if code >= 400 {
				return
			}
			ictx.funcParams[paramIdx] = *argp
		} else if argType.Kind() == reflect.Struct {
			argp := reflect.New(argType)
			var err error
			for _, parser := range ictx.httpParsers {
				err = parser(ictx.req, argType, &argp)
				if err != nil {
					nerr := nerror.NewCommonError(err.Error())
					p.sendError(nerr)
					code = http.StatusBadRequest
					body = nerr.Masking()
					return
				}
			}
			validate(&argp, &body, &code)
			if code >= 400 {
				return
			}
			ictx.funcParams[paramIdx] = argp.Elem()
		} else if argType.Kind() == reflect.Slice {
			argp := reflect.New(argType)
			var err error
			err = parseInvokejson(ictx.req, argType, &argp)
			if err != nil {
				nerr := nerror.NewCommonError(err.Error())
				p.sendError(nerr)
				code = http.StatusBadRequest
				body = nerr.Masking()
				return
			}
			validate(&argp, &body, &code)
			if code >= 400 {
				return
			}
			ictx.funcParams[paramIdx] = argp.Elem()
		} else if argType.Kind() == reflect.Ptr && argType.Elem().Kind() == reflect.Struct {
			argp := reflect.New(argType.Elem())
			var err error
			for _, parser := range ictx.httpParsers {
				err = parser(ictx.req, argType.Elem(), &argp)
				if err != nil {
					nerr := nerror.NewCommonError(err.Error())
					p.sendError(nerr)
					code = http.StatusBadRequest
					body = nerr.Masking()
					return
				}
			}
			validate(&argp, &body, &code)
			if code >= 400 {
				return
			}
			ictx.funcParams[paramIdx] = argp
		} else {
			nerr := nerror.NewCommonError("Unsupported function parameter: %s %s", argType.Name(), argType.String())
			p.sendError(nerr)
			code = http.StatusBadRequest
			body = nerr.Masking()
			return
		}
	}

	// Call "method" belonging to struct "class"
	if decorator := p.needDecorate(ictx.fn); decorator != nil {
		ictx.decorator = decorator
		callWithDecorator(ictx)
	} else {
		ictx.funcReturns = ictx.fn.Call(ictx.funcParams)
	}
	body, code = resolveFuncReturn(ictx.funcReturns)
	return
}

func (p *Invoker) invokeMethodWithStatus(ictx *InvokerContext, status int) (body interface{}, code int) {
	validate := func(argp *reflect.Value, body *interface{}, code *int) {
		if v, ok := argp.Interface().(Validator); ok {
			err := v.Validate()
			if err != nil {
				p.sendError(err)
				*code = http.StatusBadRequest
				if ne, ok := err.(nerror.NervError); ok {
					*body = ne.Masking()
				} else {
					*body = err
				}
				return
			}
		}
		return
	}
	withAuth := func(body *interface{}, code *int) *reflect.Value {
		user, clientToken := readHeader(ictx.req)
		info, err := p.Passport.CheckAuthorization(user, "", ictx.moduleName, ictx.resourceName,
			clientToken, p.Passport.GetAuthToken(ictx.moduleName), 0)
		if err != nil {
			p.sendError(err)
			*code = http.StatusBadRequest
			if ne, ok := err.(nerror.NervError); ok {
				resolveError(ne, body, code)
			} else {
				*body = err
			}
			return nil
		}
		value := setContext(user, info)

		return value
	}
	withoutAuth := func(body *interface{}, code *int) *reflect.Value {
		user, _ := readHeader(ictx.req)
		value := setContext(user, nil)

		return value
	}

	for paramIdx := 1; paramIdx < ictx.fn.Type().NumIn(); paramIdx++ {
		argType := ictx.fn.Type().In(paramIdx)
		if argType.String() == "context.Context" {
			var argp *reflect.Value
			if ictx.needAuth {
				argp = withAuth(&body, &code)
			} else {
				argp = withoutAuth(&body, &code)
			}
			if code >= 400 {
				return
			}
			ictx.funcParams[paramIdx] = *argp
		} else if argType.Kind() == reflect.Struct {
			argp := reflect.New(argType)
			var err error
			for _, parser := range ictx.httpParsers {
				err = parser(ictx.req, argType, &argp)
				if err != nil {
					nerr := nerror.NewCommonError(err.Error())
					p.sendError(nerr)
					code = http.StatusBadRequest
					body = nerr.Masking()
					return
				}
			}
			validate(&argp, &body, &code)
			if code >= 400 {
				return
			}
			ictx.funcParams[paramIdx] = argp.Elem()
		} else if argType.Kind() == reflect.Slice {
			argp := reflect.New(argType)
			var err error
			err = parseInvokejson(ictx.req, argType, &argp)
			if err != nil {
				nerr := nerror.NewCommonError(err.Error())
				p.sendError(nerr)
				code = http.StatusBadRequest
				body = nerr.Masking()
				return
			}
			validate(&argp, &body, &code)
			if code >= 400 {
				return
			}
			ictx.funcParams[paramIdx] = argp.Elem()
		} else if argType.Kind() == reflect.Ptr && argType.Elem().Kind() == reflect.Struct {
			argp := reflect.New(argType.Elem())
			var err error
			for _, parser := range ictx.httpParsers {
				err = parser(ictx.req, argType.Elem(), &argp)
				if err != nil {
					nerr := nerror.NewCommonError(err.Error())
					p.sendError(nerr)
					code = http.StatusBadRequest
					body = nerr.Masking()
					return
				}
			}
			validate(&argp, &body, &code)
			if code >= 400 {
				return
			}
			ictx.funcParams[paramIdx] = argp
		} else {
			nerr := nerror.NewCommonError("Unsupported function parameter: %s %s", argType.Name(), argType.String())
			p.sendError(nerr)
			code = http.StatusBadRequest
			body = nerr.Masking()
			return
		}
	}

	// Call "method" belonging to struct "class"
	if decorator := p.needDecorate(ictx.fn); decorator != nil {
		ictx.decorator = decorator
		callWithDecorator(ictx)
	} else {
		ictx.funcReturns = ictx.fn.Call(ictx.funcParams)
	}
	body, code = resolveFuncReturnWithStatus(ictx.funcReturns, status)
	return
}

func (p *Invoker) needDecorate(fn *reflect.Value) Decorator {
	if p.decorator != nil && p.decorateEnable[fn.Pointer()] {
		if d, ok := p.decorateOnMethod[fn.Pointer()]; ok {
			return d
		}
		return p.decorator
	}
	return nil
}

func callWithDecorator(ictx *InvokerContext) {
	err := ictx.decorator.Before(ictx)
	if err != nil {
		ictx.funcReturns[1] = reflect.ValueOf(err)
		return
	}

	ictx.funcReturns = ictx.fn.Call(ictx.funcParams)

	err = ictx.decorator.After(ictx)
	if err != nil {
		ictx.funcReturns[1] = reflect.ValueOf(err)
		return
	}
}

func invokeGetMethod(c *brick.Container, class, method string) (reflect.Value, []reflect.Value, error) {
	fn := reflect.Value{}
	params := []reflect.Value{}
	svc := c.GetByName(class)
	if svc == nil {
		return fn, params, fmt.Errorf("class %s does't exist", class)
	}
	t := reflect.TypeOf(svc)
	m, ok := t.MethodByName(method)
	if !ok {
		return fn, params, fmt.Errorf("method %s.%s doesn't exist from %v", class, method, t)
	}
	fn = m.Func
	// set "*this" to member method
	params = make([]reflect.Value, fn.Type().NumIn())
	params[0] = reflect.ValueOf(svc)
	return fn, params, nil
}

func setContext(user string, authInfo *auth.PassportAuthInfo) *reflect.Value {
	var ctx context.Context
	ctx = context.WithValue(context.Background(), rest.CurrentUserKey(), user)
	ctx = context.WithValue(ctx, orm.ContextCurrentUser(), user)
	if authInfo != nil {
		ctx = context.WithValue(ctx, rest.CurrentAuthInfoKey(), authInfo)
		ctx = context.WithValue(ctx, rest.CurrentTenantKey(), authInfo.TenantName)
	}
	value := reflect.ValueOf(ctx)
	return &value
}

func readHeader(req *http.Request) (string, string) {
	curUser := req.Header.Get(rest.GiligiliUser)
	curToken := req.Header.Get(rest.GiligiliToken)

	return curUser, curToken
}

func resolveError(ne nerror.NervError, body *interface{}, code *int) {
	switch ne.Type() {
	case nerror.CommonErrorType:
		*code = http.StatusInternalServerError
		break
	case nerror.ArgumentErrorType:
		*code = http.StatusBadRequest
		break
	case nerror.AuthorizationErrorType:
		*code = http.StatusForbidden
		break
	case nerror.PanicErrorType:
		*code = http.StatusInternalServerError
		break
	}
	*body = ne.Masking()
}

func resolveFuncReturn(fnRet []reflect.Value) (body interface{}, code int) {
	code = http.StatusOK
	if e, ok := fnRet[1].Interface().(error); ok {
		code = http.StatusInternalServerError
		if ne, ok := e.(nerror.NervError); ok {
			resolveError(ne, &body, &code)
		} else {
			body = e.Error()
		}
	} else {
		code = http.StatusOK
		body = fnRet[0].Interface()
	}
	return
}

func resolveFuncReturnWithStatus(fnRet []reflect.Value, status int) (body interface{}, code int) {
	code = http.StatusOK
	if e, ok := fnRet[1].Interface().(error); ok {
		code = http.StatusInternalServerError
		if ne, ok := e.(nerror.NervError); ok {
			resolveError(ne, &body, &code)
		} else {
			body = e.Error()
		}
	} else {
		code = status
		body = fnRet[0].Interface()
	}
	return
}

func parseInvokeQuery(req *http.Request, argType reflect.Type, argp *reflect.Value) error {
	argElem := argp.Elem()
	ec := make(chan error, 1)
	defer close(ec)
	for i := 0; i < argType.NumField(); i++ {
		tagValue, ok := argType.Field(i).Tag.Lookup("invoke-query")
		if !ok {
			continue
		}
		field := argElem.Field(i)
		switch field.Kind() {
		case reflect.String:
			argValue := getQueryParamString(req, tagValue, "")
			field.Set(reflect.ValueOf(argValue))
			ec <- nil
			break
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			argValue, err := getQueryParamInt64(req, tagValue, 0)
			v := reflect.ValueOf(argValue).Convert(field.Type())
			field.Set(v)
			ec <- err
			break
		case reflect.Slice:
			argValue := getQueryParamStringSlice(req, tagValue, []string{})
			if field.Type().Elem().Kind() == reflect.String {
				field.Set(reflect.ValueOf(argValue))
			} else if field.Type().Elem().Kind() == reflect.Interface {
				v := []interface{}{}
				for _, each := range argValue {
					v = append(v, each)
				}
				field.Set(reflect.ValueOf(v))
			}
			ec <- nil
			break
		}
		if e := <-ec; e != nil {
			e = fmt.Errorf("Parsing invalid argument: %+v", e)
			return e
		}
	}
	return nil
}

func parseInvokePath(req *http.Request, argType reflect.Type, argp *reflect.Value) error {
	argElem := argp.Elem()
	for i := 0; i < argType.NumField(); i++ {
		tagValue, ok := argType.Field(i).Tag.Lookup("invoke-path")
		if !ok {
			continue
		}
		field := argElem.Field(i)
		argValue := chi.URLParam(req, tagValue)
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			num, err := strconv.ParseInt(argValue, 10, 64)
			if err != nil {
				return err
			}
			v := reflect.ValueOf(num).Convert(field.Type())
			field.Set(v)
			break
		case reflect.String:
			field.Set(reflect.ValueOf(argValue))
			break
		}
	}
	return nil
}

func parseInvokejson(req *http.Request, argType reflect.Type, argp *reflect.Value) error {
	arg := json.RawMessage{}
	err := render.Bind(req.Body, &arg)
	if err != nil && err != io.EOF {
		return err
	}
	if len(arg) == 0 {
		return nil
	}
	err = json.Unmarshal(arg, argp.Interface())
	if err != nil {
		return err
	}
	return nil
}

func parseInvokeHeader(req *http.Request, argType reflect.Type, argp *reflect.Value) error {
	argElem := argp.Elem()
	for i := 0; i < argType.NumField(); i++ {
		tagValue, ok := argType.Field(i).Tag.Lookup("invoke-header")
		if !ok {
			continue
		}
		field := argElem.Field(i)
		header := req.Header.Get(tagValue)
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			num, err := strconv.ParseInt(header, 10, 64)
			if err != nil {
				return err
			}
			v := reflect.ValueOf(num).Convert(field.Type())
			field.Set(v)
			break
		case reflect.String:
			field.Set(reflect.ValueOf(header))
			break
		}
	}
	return nil
}
