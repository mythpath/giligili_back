package invoker_test

import (
	"context"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/nerror"
	"selfText/giligili_back/libcommon/net/http/invoker"
	"selfText/giligili_back/libcommon/net/http/rest"
	"selfText/giligili_back/libcommon/passport/auth"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/go-chi/chi"
)

type TestClass struct {
}

func (p *TestClass) Do() error {
	log.Printf("hello world!")
	return nil
}

func (p *TestClass) DoSingle(ctx context.Context) (interface{}, error) {
	str := "hello world!"
	log.Printf(str)
	return str, nil
}

type ListValues struct {
	Where    string        `invoke-query:"where"`
	Values   []interface{} `invoke-query:"values"`
	Order    string        `invoke-query:"order"`
	Page     int           `invoke-query:"page"`
	PageSize uint64        `invoke-query:"pageSize"`
}

func (p *TestClass) RunList(ctx context.Context, in ListValues) (interface{}, error) {
	s := fmt.Sprintf("Get list params: %+v", in)
	log.Println(s)
	return in, nil
}

type CreateValues struct {
	Name string `json:"name"`
	Data int    `json:"data"`
}

func (p *TestClass) RunCreate(ctx context.Context, in CreateValues) (interface{}, error) {
	s := fmt.Sprintf("Get create params: %+v", in)
	log.Println(s)
	return map[string]interface{}{"name": in.Name, "data": in.Data}, nil
}

type UpdateValues struct {
	Name string `json:"name"`
	Data int    `json:"data"`
}

func (p *TestClass) RunUpdate(ctx context.Context, in UpdateValues) (interface{}, error) {
	s := fmt.Sprintf("Get update params: %+v", in)
	log.Println(s)
	return map[string]interface{}{"name": in.Name, "data": in.Data}, nil
}

type ExecuteValues struct {
	Name   string `json:"name"`
	Action string `json:"action"`
}

func (p *TestClass) RunExecuteWithErr(ctx context.Context, in ExecuteValues) (interface{}, error) {
	s := fmt.Sprintf("Get execute params: %+v", in)
	log.Println(s)
	return map[string]interface{}{"name": in.Name}, fmt.Errorf("%s", s)
}

type GetValues struct {
	Name string `invoke-path:"name"`
	Id   int64  `invoke-path:"id"`
}

func (p *TestClass) RunGet(ctx context.Context, in GetValues) (interface{}, error) {
	s := fmt.Sprintf("Get get detail params: %+v", in)
	log.Println(s)
	return map[string]interface{}{"name": in.Name, "id": in.Id}, nil
}

type DeleteValues struct {
	Name         string        `json:"name"`
	Id           string        `invoke-path:"id"`
	errorPattern []interface{} `invoke-query:"pattern"`
}

func (p *TestClass) RunDelete(ctx context.Context, in DeleteValues) (interface{}, error) {
	s := fmt.Sprintf("Get delte params: %+v", in)
	log.Println(s)
	return map[string]interface{}{"name": in.Name, "id": in.Id}, nil
}

type FirstValues struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type SecondValues struct {
	Name string `invoke-query:"name"`
	Id   string `invoke-path:"id"`
}

func (p *TestClass) RunTriple(ctx context.Context, ina FirstValues, inb SecondValues) (interface{}, error) {
	s := fmt.Sprintf("Get two params: %+v, %+v", ina, inb)
	log.Println(s)
	return map[string]interface{}{
		"first name":  ina.Name,
		"first id":    ina.Id,
		"second name": inb.Name,
		"second id":   inb.Id}, nil
}

type ValuesWithArgError struct {
	Name string `invoke-path:"name"`
}

func (p *ValuesWithArgError) Validate() error {
	if len(p.Name) > 1 {
		nerr := nerror.NewArgumentError("in")
		nerr.FieldError("name", fmt.Sprintf("Length of name [%s] exceeds 1", p.Name))
		return nerr
	}
	return nil
}

func (p *TestClass) RunWithArgError(ctx context.Context, in ValuesWithArgError) (interface{}, error) {
	s := fmt.Sprintf("Get RunWithArgError params: %+v", in)
	log.Println(s)
	return map[string]interface{}{"name": in.Name}, nerror.NewCommonError("test error")
}

type TextArg struct {
	Name string `invoke-path:"name"`
}

func (p *TestClass) ReturnText(ctx context.Context, in TextArg) ([]byte, error) {
	text := fmt.Sprintf("name: %s\n", in.Name)
	return []byte(text), nil
}

type HeaderArg struct {
	Name string `invoke-header:"PRIVATE-NAME"`
}

func (p *TestClass) GetHeader(ctx context.Context, in HeaderArg) (interface{}, error) {
	log.Printf("name: %s\n", in.Name)
	return in, nil
}

type PointerArg struct {
	Name string `json:"name"`
	ID   int    `invoke-path:"id"`
}

func (p *TestClass) GetByPointer(ctx context.Context, in *PointerArg) (interface{}, error) {
	log.Printf("name: %s, id: %d\n", in.Name, in.ID)
	return in, nil
}

type TestExecuter struct {
	Invoker *invoker.Invoker `inject:"test-invoker"`
	Class   *TestClass       `inject:"test-class"`
}

// go test -v -run TestInvoker1
func TestInvoker1(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	c := brick.NewContainer()
	c.Add(&invoker.Invoker{}, "test-invoker", nil)
	c.Add(&TestClass{}, "test-class", nil)
	c.Add(&TestExecuter{}, "test-exec", nil)
	c.Add(&brick.JSONConfigService{}, "config", brick.FactoryFunc(func() interface{} {
		return brick.NewJSONConfigService("config/test_config.json")
	}))
	c.Add(&auth.PassportClient{}, "passport-client", nil)
	c.Build()

	s := c.GetByName("test-exec").(*TestExecuter)
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		// panic case, uncomment it to test
		// r.Get("/do", s.Invoker.Execute("test-class", "Do"))

		// curl -i -H "NERV-USER: admin" 'localhost:23333/do-single'
		r.Get("/do-single", s.Invoker.Execute("test-class", "DoSingle"))

		// curl -i -H "NERV-USER: admin" 'localhost:23333/list?where=name&values=obj&order=test&page=1&pageSize=10'
		r.Get("/list", s.Invoker.Execute("test-class", "RunList"))

		// curl -i -H "NERV-USER: admin" 'localhost:23333/get-detail/admin/666'
		r.Get("/get-detail/{name}/{id}", s.Invoker.Execute("test-class", "RunGet"))

		// curl -i -H "NERV-USER: admin" -X POST -d '{"name": "admin", "data": 42}' 'localhost:23333/create'
		r.Post("/create", s.Invoker.Execute("test-class", "RunCreate"))

		// curl -i -H "NERV-USER: admin" -X PUT -d '{"name": "anonymous", "data": -1}' 'localhost:23333/'
		r.Put("/", s.Invoker.Execute("test-class", "RunUpdate"))

		// test runtime panic error
		// curl -i -H "NERV-USER: admin" -X DELETE -d '{"name": "admin"}' 'localhost:23333/777'
		r.Delete("/{id}", s.Invoker.Execute("test-class", "RunDelete"))

		// test error return
		// curl -i -H "NERV-USER: admin" -X POST -d '{"name": "admin", "action": "sleep"}' 'localhost:23333/do-something'
		r.Post("/do-something", s.Invoker.Execute("test-class", "RunExecuteWithErr"))

		// mixed tag
		// curl -i -H "NERV-USER: admin" -X PUT -d '{"name": "admin", "id": "0xff"}' 'localhost:23333/do-triple/42?name=anonymous'
		r.Put("/do-triple/{id}", s.Invoker.Execute("test-class", "RunTriple"))

		// arg error test
		// curl -i -H "NERV-USER: admin" localhost:23333/err/admin
		r.Get("/err/{name}", s.Invoker.Execute("test-class", "RunWithArgError"))

		// return text
		// curl -i -H "NERV-USER: admin" localhost:23333/text/dingcloud
		r.Get("/text/{name}", s.Invoker.ExecuteText("test-class", "ReturnText"))

		// test invoke-header tag
		// curl -i -H "NERV-USER: admin" -H "PRIVATE-NAME: admin-test" localhost:23333/header
		r.Get("/header", s.Invoker.Execute("test-class", "GetHeader"))

		// test pointer type arg
		// curl -i -H "NERV-USER: admin" -X POST -d '{"name": "admin"}' localhost:23333/pointer/12
		r.Post("/pointer/{id}", s.Invoker.Execute("test-class", "GetByPointer"))
	})
	http.ListenAndServe(":23333", r)
}

func (p *TestClass) GetContext(ctx context.Context) (interface{}, error) {
	user := invoker.NervCtxGetUser(ctx)
	info := invoker.NervCtxGetAuthInfo(ctx)
	tenant := invoker.NervCtxGetTenant(ctx)
	log.Printf("user: %s\n", user)

	buf, _ := json.MarshalIndent(info, "", "  ")
	log.Printf("info: %s\n", string(buf))
	log.Printf("tenant: %s", tenant)
	return nil, nil
}

// go test -v -run TestInvoker2, test passport auth
func TestInvoker2(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	c := brick.NewContainer()
	c.Add(&TestClass{}, "test-class", nil)
	c.Add(&TestExecuter{}, "test-exec", nil)
	c.Add(&brick.JSONConfigService{}, "config", brick.FactoryFunc(func() interface{} {
		return brick.NewJSONConfigService("config/test_config.json")
	}))
	c.Add(&auth.PassportClient{}, "passport-client", brick.FactoryFunc(func() interface{} {
		return auth.NewPassportClient("config/static_permission.json")
	}))
	c.Add(&invoker.Invoker{}, "test-invoker", nil)
	c.Build()

	s := c.GetByName("test-exec").(*TestExecuter)
	r := chi.NewRouter()

	passport := c.GetByName("passport-client").(*auth.PassportClient)
	err := passport.RegisterOperation()
	if err != nil {
		log.Printf("register: %s", err.Error())
		return
	}
	log.Printf("token: %s", passport.GetAuthToken("test-module"))
	r.Route("/", func(r chi.Router) {
		// curl -i -H "NERV-USER: admin" -H "NERV-TOKEN: bb5c01a4edb7df4bf0a01a975d527778" localhost:23333/context
		r.Get("/context", s.Invoker.ExecuteWithAuth("test-class", "GetContext", "test-module", "1"))
		// curl -i -H "NERV-USER: admin" -H "NERV-TOKEN: a314fe31f84a1d86b84e31450d666666" -X DELETE localhost:23333/context
		r.Delete("/context", s.Invoker.ExecuteWithAuth("test-class", "GetContext", auth.PassportModule, auth.PassportResource))
	})
	http.ListenAndServe(":23333", r)
}

type MyDecorator struct {
	invoker.VirtualDecorator
}

func (p *MyDecorator) Before(ictx *invoker.InvokerContext) error {
	log.Printf("before function")
	log.Printf("context: %#v", invoker.GetFuncParamNervCtx(ictx))

	return nil
}

func (p *MyDecorator) After(ictx *invoker.InvokerContext) error {
	log.Printf("after function")
	for i, fnOut := range invoker.GetFuncReturns(ictx) {
		log.Printf("ojbect[%d] is %#v", i, fnOut)
	}
	return nil
}

func (p *MyDecorator) GetError(ictx *invoker.InvokerContext) error {
	return nil
}

type SpecialDecorator struct {
	invoker.VirtualDecorator
}

func (p *SpecialDecorator) Before(ictx *invoker.InvokerContext) error {
	log.Printf("SpecialDecorator: %s", invoker.FuncParamGetAction(invoker.GetFuncParamByIndex(ictx, 2)))

	return nil
}

func (p *SpecialDecorator) After(ictx *invoker.InvokerContext) error {
	ok, _ := invoker.CompareCallingFunction(ictx, "test-class", "SpecialDecoratingV2")
	if ok {
		return nerror.NewCommonError("SpecialDecoratingV2 error")
	}
	return nil
}

func (p *TestClass) DoDecorating(ctx context.Context) (interface{}, error) {
	user := ctx.Value(rest.CurrentUserKey())

	log.Printf("user: %s", user)
	return user, nil
}

func (p *TestClass) DoWithoutDecorating(ctx context.Context) (interface{}, error) {
	user := ctx.Value(rest.CurrentUserKey())

	log.Printf("user: %s", user)
	return user, nil
}

type specialDecoratingArg struct {
	invoker.ServiceMethodArgBase
	Num int `json:"num"`
}

func (p *specialDecoratingArg) Action() string {
	return "special action"
}

func (p *TestClass) SpecialDecorating(ctx context.Context, arg *specialDecoratingArg) (interface{}, error) {
	user := ctx.Value(rest.CurrentUserKey())

	log.Printf("user: %s, arg: %#v", user, arg)
	return user, nil
}

type specialDecoratingV2Return struct {
	User string
}

func (p *TestClass) SpecialDecoratingV2(ctx context.Context, arg *specialDecoratingArg) (interface{}, error) {
	user := invoker.NervCtxGetUser(ctx)

	log.Printf("user: %s, arg: %#v", user, arg)
	ret := &specialDecoratingV2Return{User: user}
	return ret, nil
}

// go test -v -run TestInvoker3, test decorator
func TestInvoker3(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	c := brick.NewContainer()
	c.Add(&invoker.Invoker{}, "test-invoker", nil)
	c.Add(&TestClass{}, "test-class", nil)
	c.Add(&TestExecuter{}, "test-exec", nil)
	c.Add(&brick.JSONConfigService{}, "config", brick.FactoryFunc(func() interface{} {
		return brick.NewJSONConfigService("config/test_config.json")
	}))
	c.Add(&auth.PassportClient{}, "passport-client", nil)
	c.Build()

	s := c.GetByName("test-exec").(*TestExecuter)
	r := chi.NewRouter()

	s.Invoker.SetDecorator(&MyDecorator{})

	r.Route("/", func(r chi.Router) {
		// basic docorating
		// curl -i -H "NERV-USER: admin" localhost:23333/DoDecorating
		r.Get("/DoDecorating", s.Invoker.Execute("test-class", "DoDecorating"))

		// disable decorating
		// curl -i -H "NERV-USER: admin" localhost:23333/DoWithoutDecorating
		r.Get("/DoWithoutDecorating", s.Invoker.Execute("test-class", "DoWithoutDecorating"))
		s.Invoker.DisableDecoratorOn("test-class", "DoWithoutDecorating")

		// use special decorator
		// curl -i -H "NERV-USER: admin" -X POST -d '{"num": 10}' localhost:23333/SpecialDecorating
		r.Post("/SpecialDecorating", s.Invoker.Execute("test-class", "SpecialDecorating"))
		s.Invoker.SetDecoratorOn("test-class", "SpecialDecorating", &SpecialDecorator{})

		// return error from After
		// curl -i -H "NERV-USER: admin" -X POST -d '{"num": 10}' localhost:23333/SpecialDecoratingV2
		r.Post("/SpecialDecoratingV2", s.Invoker.Execute("test-class", "SpecialDecoratingV2"))
		s.Invoker.SetDecoratorOn("test-class", "SpecialDecoratingV2", &SpecialDecorator{})
	})
	http.ListenAndServe(":23333", r)
}

type errorLogger struct {
	err chan error
}

func newErrorLogger() *errorLogger {
	p := &errorLogger{
		err: make(chan error),
	}
	go p.mainLoop()
	return p
}

func (p *errorLogger) ErrorChan() chan error {
	return p.err
}

func (p *errorLogger) mainLoop() {
	for err := range p.err {
		log.Printf("get error: %s", err.Error())
	}
}

// go test -v -run TestInvoker4, test invoker error channel
func TestInvoker4(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	errLogger := newErrorLogger()
	c := brick.NewContainer()
	c.Add(&invoker.Invoker{}, "test-invoker", brick.FactoryFunc(func() interface{} {
		return invoker.NewInvoker(nil, errLogger.ErrorChan())
	}))
	c.Add(&TestClass{}, "test-class", nil)
	c.Add(&TestExecuter{}, "test-exec", nil)
	c.Add(&brick.JSONConfigService{}, "config", brick.FactoryFunc(func() interface{} {
		return brick.NewJSONConfigService("config/test_config.json")
	}))
	c.Add(&auth.PassportClient{}, "passport-client", nil)
	c.Build()

	s := c.GetByName("test-exec").(*TestExecuter)
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		// arg error test
		// curl -i -H "NERV-USER: admin" localhost:23333/err/admin
		r.Get("/err/{name}", s.Invoker.Execute("test-class", "RunWithArgError"))

		// test runtime panic error
		// curl -i -H "NERV-USER: admin" -X DELETE -d '{"name": "admin"}' 'localhost:23333/777'
		r.Delete("/{id}", s.Invoker.Execute("test-class", "RunDelete"))
	})
	http.ListenAndServe(":23333", r)
}
