package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"runtime/debug"
	"strings"

	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/net/http/rest/render"
	"selfText/giligili_back/libcommon/orm"

	"github.com/go-chi/chi"
)

func handlePanic(w http.ResponseWriter, req *http.Request) {
	if r := recover(); r != nil {
		fmt.Printf("panic: %v\n", r)
		debug.PrintStack()
		render.Status(req, 500)
		render.JSON(w, req, r)
	}
}

// RestController
type RestController struct {
	brick.Trigger
	DB            orm.DBService     `inject:"DB"`
	ModelRegistry orm.ModelRegistry `inject:"DB"`
	container     *brick.Container
	ormService    orm.Repository
}

func (p *RestController) Init() error {
	if p.ormService == nil {
		p.ormService = &orm.OrmService{DB: p.DB, ModelRegistry: p.ModelRegistry} //newormService(p.DB, p.ModelRegistry)
	}

	return nil
}

func (p *RestController) List(w http.ResponseWriter, req *http.Request) {
	defer handlePanic(w, req)

	/**
	  获得所有查询参数
	*/
	class := chi.URLParam(req, "class")
	where := req.URL.Query().Get("where")
	selectQuery := req.URL.Query().Get("select")

	/**
	  如果select参数不为空, 则获得要查询的字段集合
	*/
	var selectFields []string
	if selectQuery != "" {
		for _, arg := range strings.Split(selectQuery, ",") {
			selectFields = append(selectFields, arg)
		}
	}

	/**
	  如果where参数不为空, 则获得values数组
	*/
	var whereValues []interface{}
	if where != "" {
		values := req.URL.Query().Get("values")
		if values == "" {
			render.Status(req, 400)
			render.JSON(w, req, fmt.Sprintf("the values query param must be provided if the where query param is exists"))
			return
		}
		for _, arg := range strings.Split(values, ",") {
			whereValues = append(whereValues, arg)
		}
	}

	order := req.URL.Query().Get("order")
	page, err := getQueryParamInt(req, "page", 0)
	if err != nil {
		render.Status(req, 400)
		render.JSON(w, req, err.Error())
		return
	}
	pageSize, err := getQueryParamInt(req, "pageSize", 10)
	if err != nil {
		render.Status(req, 400)
		render.JSON(w, req, err.Error())
		return
	}

	data, err := p.ormService.List(class, selectFields, where, whereValues, order, page, pageSize)
	if err != nil {
		render.Status(req, 500)
		render.JSON(w, req, err.Error())
		return
	}

	render.Status(req, 200)
	render.JSON(w, req, data)

	// md := p.ModelRegistry.Get(class)
	// if md == nil {
	// 	render.Status(req, 400)
	// 	render.JSON(w, req, fmt.Sprintf("class %s isn't exists", class))
	// 	return
	// }

	// //count
	// d := p.DB.GetDB().Model(md.NewSlice())
	// var count int64
	// if where != "" {
	// 	d = d.Where(where, args...)
	// }
	// if selectQuery != "" {
	// 	d = d.Select(selectFields)
	// }
	// if err := d.Count(&count).Error; err != nil {
	// 	render.Status(req, 500)
	// 	render.JSON(w, req, err.Error())
	// 	return
	// }
	// //order page
	// var page, pageCount, limit int64
	// limit = 10
	// var err error
	// paramPage := req.URL.Query().Get("page")
	// paramSize := req.URL.Query().Get("pageSize")

	// if paramSize != "" {
	// 	limit, err = strconv.ParseInt(paramSize, 10, 32)
	// 	if err != nil {
	// 		limit = 10
	// 	}
	// }

	// pageCount = (count + limit - 1) / limit

	// if paramPage != "" {
	// 	page, err = strconv.ParseInt(paramPage, 10, 32)
	// 	if err != nil {
	// 		page = 0
	// 	}
	// }

	// if page < 0 {
	// 	page = 0
	// }

	// if page >= pageCount {
	// 	page = pageCount - 1
	// }

	// if where != "" {
	// 	d = p.DB.GetDB().Where(where, args...)
	// }
	// if selectQuery != "" {
	// 	log.Printf("selectFields : %v", selectFields)
	// 	d = d.Select(selectFields)
	// }

	// order := req.URL.Query().Get("order")
	// if order != "" {
	// 	d = d.Order(order)
	// }
	// data := md.NewSlice()
	// if d.Offset(page * limit).Limit(limit).Find(data).RecordNotFound() {
	// 	render.Status(req, 200)
	// 	render.JSON(w, req, data)
	// 	return
	// }

	// render.Status(req, 200)
	// render.JSON(w, req, map[string]interface{}{"data": data, "page": page, "pageSize": limit, "pageCount": pageCount})
}

// get one obj. query params: assocations=a,b...
func (p *RestController) Get(w http.ResponseWriter, req *http.Request) {
	defer handlePanic(w, req)
	// waitTime := chi.URLParam(req, "waitTime")
	// if strings.Trim(waitTime, "") == "" {
	p.get(w, req)
	// } else {
	// 	p.getWait(w, req)
	// }
}

func (p *RestController) get(w http.ResponseWriter, req *http.Request) {
	class := chi.URLParam(req, "class")
	id := chi.URLParam(req, "id")
	ass := req.URL.Query().Get("associations")

	data, err := p.ormService.Get(class, id, ass)
	if err != nil {
		render.Status(req, 500)
		render.JSON(w, req, err.Error())
		return
	}

	render.Status(req, 200)
	render.JSON(w, req, data)

	// md := p.ModelRegistry.Get(class)
	// if md == nil {
	// 	render.Status(req, 400)
	// 	render.JSON(w, req, fmt.Sprintf("class %s isn't exists", class))
	// 	return
	// }

	// data := md.New()
	// d := p.DB.GetDB()
	// for _, as := range strings.Split(ass, ",") {
	// 	d = d.Preload(as)
	// }

	// if d.First(data, id).RecordNotFound() {
	// 	render.Status(req, 200)
	// 	render.JSON(w, req, nil)
	// 	return
	// }

	// render.Status(req, 200)
	// render.JSON(w, req, data)
}

// func (p *RestController) getWait(w http.ResponseWriter, req *http.Request) {
// 	waitKey, err := getQueryParamInt64(req, "waitKey", 0)
// 	if err != nil {
// 		render.Status(req, 400)
// 		render.JSON(w, req, err.Error())
// 		return
// 	}
// 	waitTime, err := getQueryParamInt(req, "waitTime", 30)
// 	if err != nil {
// 		render.Status(req, 400)
// 		render.JSON(w, req, err.Error())
// 		return
// 	}

// 	className := chi.URLParam(req, "class")
// 	ass := req.URL.Query().Get("associations")
// 	id, err := getURLParamUint(req, "id")
// 	if err != nil {
// 		render.Status(req, 400)
// 		render.JSON(w, req, err.Error())
// 		return
// 	}
// 	var last *WatchEvent
// 	leaseKey, events := p.ormService.WatchObject(waitKey, waitTime, className, id, ass)
// 	if waitKey == 0 {
// 		// Return leaseKey first get,then the client must call getWait use the leaseKey for the next change immediately
// 		data, err := p.ormService.Get(className, id, ass)
// 		if err != nil {
// 			render.Status(req, 500)
// 			render.JSON(w, req, err.Error())
// 			return
// 		}

// 		ormEvent := NewOrmEvent(orm_read, class(className), data)
// 		last = &WatchEvent{WaitKey: leaseKey, OrmEvent: ormEvent}
// 		render.Status(req, 200)
// 		render.JSON(w, req, last)
// 	} else {
// 		for event := range events {
// 			last = event
// 		}

// 		if last != nil {
// 			render.Status(req, 200)
// 			render.JSON(w, req, last)
// 		} else {
// 			render.Status(req, 304)
// 			render.JSON(w, req, nil)
// 		}
// 	}
// }

func (p *RestController) Create(w http.ResponseWriter, req *http.Request) {
	p.CreateCtx(context.TODO(), w, req)
}

func (p *RestController) CreateCtx(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	defer handlePanic(w, req)

	class := chi.URLParam(req, "class")
	md := p.ModelRegistry.Get(class)
	if md == nil {
		render.Status(req, 400)
		render.JSON(w, req, fmt.Sprintf("class %s isn't exists", class))
		return
	}

	data := md.New()
	if err := render.Bind(req.Body, data); err != nil {
		render.Status(req, 400)
		render.JSON(w, req, err.Error())
		return
	}

	if err := p.ormService.CreateCtx(p.contextWithUser(ctx, req), class, data); err != nil {
		render.Status(req, 500)
		render.JSON(w, req, err.Error())
		return
	}

	// if err := p.DB.GetDB().Create(data).Error; err != nil {
	// 	render.Status(req, 500)
	// 	render.JSON(w, req, err.Error())
	// 	return
	// }

	render.Status(req, 200)
	render.JSON(w, req, data)

	p.raise(fmt.Sprintf("%s.Create", class), data)
}

func (p *RestController) Remove(w http.ResponseWriter, req *http.Request) {
	p.RemoveCtx(context.TODO(), w, req)
}

func (p *RestController) RemoveCtx(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	defer handlePanic(w, req)

	class := chi.URLParam(req, "class")
	id := chi.URLParam(req, "id")

	md := p.ModelRegistry.Get(class)
	if md == nil {
		render.Status(req, 400)
		render.JSON(w, req, fmt.Sprintf("class %s isn't exists", class))
		return
	}

	data, err := p.ormService.RemoveCtx(p.contextWithUser(ctx, req), class, id, false)
	if err != nil {
		render.Status(req, 500)
		render.JSON(w, req, err.Error())
		return
	}

	// data := md.New()
	// if err := p.DB.GetDB().First(data, id).Error; err != nil {
	// 	render.Status(req, 400)
	// 	render.JSON(w, req, err.Error())
	// 	return
	// }

	// if err := p.DB.GetDB().Unscoped().Delete(data).Error; err != nil {
	// 	render.Status(req, 500)
	// 	render.JSON(w, req, err.Error())
	// 	return
	// }

	render.Status(req, 200)
	p.raise(fmt.Sprintf("%v.Delete", class), data)
}

func (p *RestController) Update(w http.ResponseWriter, req *http.Request) {
	p.UpdateCtx(context.TODO(), w, req)
}

func (p *RestController) UpdateCtx(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	defer handlePanic(w, req)

	class := chi.URLParam(req, "class")
	md := p.ModelRegistry.Get(class)
	if md == nil {
		render.Status(req, 400)
		render.JSON(w, req, fmt.Sprintf("class %s isn't exists", class))
		return
	}

	data := md.New()
	if err := render.Bind(req.Body, data); err != nil {
		render.Status(req, 400)
		render.JSON(w, req, err.Error())
		return
	}

	if err := p.ormService.UpdateCtx(p.contextWithUser(ctx, req), class, data); err != nil {
		render.Status(req, 500)
		render.JSON(w, req, err.Error())
		return
	}

	// if err := p.DB.GetDB().Save(data).Error; err != nil {
	// 	render.Status(req, 500)
	// 	render.JSON(w, req, err.Error())
	// 	return
	// }

	render.Status(req, 200)
	render.JSON(w, req, data)
	p.raise(fmt.Sprintf("%v.Update", class), data)
}

// InvokeServiceFunc call the  func of giligili
//  Transfer user:
//  1. set http head GILIGILI-USER in the request
//  2. define first argument as context.Context in the func. e.g.
//  3. get current user from the context
//  e.g.
//     func (p *foo) Do(ctx context.Context, ...) error {
//         user := ctx.Value(rest.CurrentUserKey()))
//     }
func (p *RestController) InvokeServiceFunc() func(w http.ResponseWriter, req *http.Request) {
	return func(c *brick.Container) func(w http.ResponseWriter, req *http.Request) {
		return func(w http.ResponseWriter, req *http.Request) {
			invokeService(c, w, req)
		}
	}(p.container)
}

func invokeService(c *brick.Container, w http.ResponseWriter, req *http.Request) {
	defer handlePanic(w, req)

	class := chi.URLParam(req, "class")
	methodName := chi.URLParam(req, "id")
	curUser := req.Header.Get(GiligiliUser)

	svc := c.GetByName(class)
	if svc == nil {
		render.Status(req, 404)
		render.JSON(w, req, fmt.Sprintf("giligili %s isn't exists", class))
		return
	}

	t := reflect.TypeOf(svc)
	if m, b := t.MethodByName(methodName); b != true {
		render.Status(req, 404)
		render.JSON(w, req, fmt.Sprintf("method %s.%s isn't exists.from %v", class, methodName, t))
		return

	} else {
		args := []json.RawMessage{}
		err := render.Bind(req.Body, &args)
		if err != nil && err != io.EOF {
			render.Status(req, 400)
			render.JSON(w, req, err.Error())
			return
		}

		in := []reflect.Value{reflect.ValueOf(svc)}
		funcType := m.Func.Type()
		if funcType.NumIn() > 1 {
			argType := funcType.In(1)
			var ctx context.Context
			step := 1
			if argType.Name() == "Context" {
				if curUser != "" {
					ctx = context.WithValue(context.Background(), CurrentUserKey(), curUser)
					ctx = context.WithValue(ctx, orm.ContextCurrentUser(), curUser)
					in = append(in, reflect.ValueOf(ctx))
					step = 2
				} else {
					render.Status(req, 400)
					render.JSON(w, req, fmt.Sprintf("giligili requires context,but current user is empty"))
					return
				}
			}

			for i, arg := range args {
				argType := funcType.In(i + step)
				argValue := reflect.New(argType)
				if err := json.Unmarshal(arg, argValue.Interface()); err == nil {
					in = append(in, argValue.Elem())
				} else {
					render.Status(req, 500)
					render.JSON(w, req, err.Error())
					return
				}
			}
		}

		values := m.Func.Call(in)
		ret := []interface{}{}
		httpCode := 200
		for _, value := range values {
			rawValue := value.Interface()
			if e, ok := rawValue.(error); ok {
				httpCode = 500
				ret = append(ret, e.Error())
			} else {
				ret = append(ret, rawValue)
			}

		}
		render.Status(req, httpCode)
		render.JSON(w, req, ret)
	}
}

// InvokeServiceRawMessageFunc call the  func of giligili with json.RawMessage
//  Transfer user:
//  1. set http head GILIGILI-USER in the request
//  2. define first argument as context.Context in the func. e.g.
//  3. get current user from the context
//  e.g.
//     func (p *foo) Do(ctx context.Context, msg json.RawMessage) (string,error) {
//         user := ctx.Value(rest.CurrentUserKey()))
//     }
func (p *RestController) InvokeServiceRawMessageFunc() func(w http.ResponseWriter, req *http.Request) {
	return func(c *brick.Container) func(w http.ResponseWriter, req *http.Request) {
		return func(w http.ResponseWriter, req *http.Request) {
			invokeServiceRawMessage(c, w, req)
		}
	}(p.container)
}

func invokeServiceRawMessage(c *brick.Container, w http.ResponseWriter, req *http.Request) {
	defer handlePanic(w, req)

	class := chi.URLParam(req, "class")
	methodName := chi.URLParam(req, "id")
	curUser := req.Header.Get(GiligiliUser)

	svc := c.GetByName(class)
	if svc == nil {
		render.Status(req, 404)
		render.JSON(w, req, fmt.Sprintf("giligili %s isn't exists", class))
		return
	}

	t := reflect.TypeOf(svc)
	if m, b := t.MethodByName(methodName); b != true {
		render.Status(req, 404)
		render.JSON(w, req, fmt.Sprintf("method %s.%s isn't exists.from %v", class, methodName, t))
		return

	} else {
		arg := json.RawMessage{}
		if err := render.Bind(req.Body, &arg); err != nil {
			render.Status(req, 400)
			render.JSON(w, req, err.Error())
			return
		}

		in := []reflect.Value{reflect.ValueOf(svc)}
		funcType := m.Func.Type()
		if funcType.NumIn() > 1 {
			argType := funcType.In(1)
			var ctx context.Context

			if argType.Name() == "Context" {
				if curUser != "" {
					ctx = context.WithValue(context.Background(), CurrentUserKey(), curUser)
					ctx = context.WithValue(ctx, orm.ContextCurrentUser(), curUser)
					in = append(in, reflect.ValueOf(ctx))
				} else {
					render.Status(req, 400)
					render.JSON(w, req, fmt.Sprintf("giligili requires context,but current user is empty"))
					return
				}
			}
			in = append(in, reflect.ValueOf(arg))
		}

		values := m.Func.Call(in)
		ret := []interface{}{}
		httpCode := 200
		if len(values) != 2 {
			render.Status(req, httpCode)
			render.JSON(w, req, fmt.Sprintf("giligili must return 2 values, (string,error)"))
			return
		}

		for _, value := range values {
			rawValue := value.Interface()
			if e, ok := rawValue.(error); ok {
				httpCode = 500
				ret = append(ret, e.Error())
			} else {
				ret = append(ret, rawValue)
			}
		}
		render.Status(req, httpCode)
		render.JSON(w, req, ret)
	}
}

func (p *RestController) InvokeObj(w http.ResponseWriter, req *http.Request) {
	defer handlePanic(w, req)

	class := chi.URLParam(req, "class")
	id := chi.URLParam(req, "id")
	methodName := chi.URLParam(req, "method")

	md := p.ModelRegistry.Get(class)
	if md == nil {
		render.Status(req, 404)
		render.JSON(w, req, fmt.Sprintf("class %s isn't exists", class))
		return
	}

	data := md.New()
	if err := p.DB.GetDB().First(data, id).Error; err != nil {
		render.Status(req, 404)
		render.JSON(w, req, err.Error())
		return
	}

	t := reflect.TypeOf(data)
	if m, b := t.MethodByName(methodName); b != true {
		render.Status(req, 404)
		render.JSON(w, req, fmt.Sprintf("%s/%s/%s isn't exists", class, id, methodName))
		return

	} else {
		args := []interface{}{}
		if err := render.Bind(req.Body, &args); err != nil {
			render.Status(req, 400)
			render.JSON(w, req, err.Error())
			return
		}

		in := []reflect.Value{reflect.ValueOf(data)}
		for _, arg := range args {
			in = append(in, reflect.ValueOf(arg))
		}
		values := m.Func.Call(in)
		ret := []interface{}{}
		httpCode := 200
		for _, value := range values {
			rawValue := value.Interface()
			if e, ok := rawValue.(error); ok {
				httpCode = 500
				ret = append(ret, e.Error())
			} else {
				ret = append(ret, rawValue)
			}

		}
		render.Status(req, httpCode)
		render.JSON(w, req, ret)
	}
}

func (p *RestController) SetContainer(c *brick.Container) {
	log.Printf("SetContainer:%+v\n", c)
	p.container = c

	log.Printf("SetContainer:%+v\n", p.container)
}

func (p *RestController) raise(event string, data interface{}) {
	p.Emmit(event, data)
}

func (p *RestController) contextWithUser(parent context.Context, req *http.Request) context.Context {
	curUser := req.Header.Get(GiligiliUser)
	return context.WithValue(parent, orm.ContextCurrentUser(), curUser)
}
