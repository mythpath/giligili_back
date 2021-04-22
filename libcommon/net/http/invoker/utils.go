package invoker

import (
	"context"
	"selfText/giligili_back/libcommon/nerror"
	"selfText/giligili_back/libcommon/net/http/rest"
	"selfText/giligili_back/libcommon/passport/auth"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// If the arg of one method implements Validator, it will be validated in customized way
type Validator interface {
	Validate() error
}

type ServiceMethodArg interface {
	Validator
	Action() string
	Catalogue() interface{}
}

type ServiceMethodArgBase struct {
}

func (p *ServiceMethodArgBase) Validate() error        { return nil }
func (p *ServiceMethodArgBase) Action() string         { return "" }
func (p *ServiceMethodArgBase) Catalogue() interface{} { return nil }

func getQueryParamString(req *http.Request, name string, def string) string {
	qs := req.URL.Query().Get(name)
	if qs == "" {
		return def
	}
	return qs
}

func getQueryParamStringSlice(req *http.Request, name string, def []string) []string {
	qs := []string{}
	v := req.URL.Query().Get(name)
	if v == "" {
		return def
	}
	for _, arg := range strings.Split(v, ",") {
		arg = strings.TrimSpace(arg)
		qs = append(qs, arg)
	}
	return qs
}

func getQueryParamInt64(req *http.Request, name string, def int64) (int64, error) {
	qs := req.URL.Query().Get(name)
	if qs == "" {
		return def, nil
	}
	value, err := strconv.ParseInt(qs, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("%s must be an int64,but it's %v", name, qs)
	}
	return int64(value), nil
}

func NervCtxGetUser(ctx context.Context) string {
	res, ok := ctx.Value(rest.CurrentUserKey()).(string)
	if !ok {
		return ""
	}
	return res
}

func NervCtxGetAuthInfo(ctx context.Context) *auth.PassportAuthInfo {
	res, ok := ctx.Value(rest.CurrentAuthInfoKey()).(*auth.PassportAuthInfo)
	if !ok {
		return nil
	}
	return res
}

func NervCtxGetTenant(ctx context.Context) string {
	res, ok := ctx.Value(rest.CurrentTenantKey()).(string)
	if !ok {
		return ""
	}
	return res
}

func GetModuleName(ictx *InvokerContext) string {
	return ictx.moduleName
}

func GetResourceName(ictx *InvokerContext) string {
	return ictx.resourceName
}

func GetHttpRequest(ictx *InvokerContext) *http.Request {
	return ictx.req
}

func GetFuncParamNervCtx(ictx *InvokerContext) context.Context {
	for i := range ictx.funcParams {
		if ctx, ok := ictx.funcParams[i].Interface().(context.Context); ok {
			return ctx
		}
	}
	return nil
}

func GetFuncParamByIndex(ictx *InvokerContext, i int) interface{} {
	// first method arg is 1, because 0th arg is for *this object
	if i < 1 || i >= len(ictx.funcParams) {
		return nil
	}
	return ictx.funcParams[i].Interface()
}

func GetCallingFunction(ictx *InvokerContext) interface{} {
	return ictx.fn.Interface()
}

func CompareCallingFunction(ictx *InvokerContext, class, method string) (bool, error) {
	fn, _, err := invokeGetMethod(ictx.container, class, method)
	if err != nil {
		return false, err
	}
	if ictx.fn.Pointer() != fn.Pointer() {
		return false, nerror.NewCommonError("[%#v] : [%#v] not match", ictx.fn, fn)
	}
	return true, nil
}

const InvalidProjectName = "\n\ti\tn\tv\ta\tl\ti\td\n"

func FuncParamGetProjectName(obj interface{}) string {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Struct {
		return InvalidProjectName
	}
	fieldName := []string{"ProjectName"}
	var projectName reflect.Value
	for i := range fieldName {
		projectName = v.FieldByName(fieldName[i])
		if projectName.Kind() == reflect.String {
			return projectName.Interface().(string)
		}
	}
	return InvalidProjectName
}

const InvalidProjectID int64 = -1

func FuncParamGetProjectID(obj interface{}) int64 {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Struct {
		return InvalidProjectID
	}
	fieldName := []string{"ProjectID"}
	var projectID reflect.Value
	for i := range fieldName {
		projectID = v.FieldByName(fieldName[i])
		switch projectID.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return projectID.Convert(reflect.TypeOf(int64(0))).Interface().(int64)
		case reflect.String:
			num, err := strconv.ParseInt(projectID.Interface().(string), 10, 64)
			if err != nil {
				return InvalidProjectID
			}
			return num
		}
	}
	return InvalidProjectID
}

func FuncParamGetAction(obj interface{}) string {
	arg, ok := obj.(ServiceMethodArg)
	if !ok {
		return ""
	}
	return arg.Action()
}

func FuncParamGetCatalogue(obj interface{}) interface{} {
	arg, ok := obj.(ServiceMethodArg)
	if !ok {
		return nil
	}
	return arg.Catalogue()
}

func GetError(ictx *InvokerContext) error {
	for i := range ictx.funcReturns {
		if err, ok := ictx.funcReturns[i].Interface().(error); ok {
			return err
		}
	}
	return nil
}

func GetFuncReturns(ictx *InvokerContext) []interface{} {
	elememts := make([]interface{}, len(ictx.funcReturns))
	for i := range ictx.funcReturns {
		elememts[i] = ictx.funcReturns[i].Interface()
	}
	return elememts
}
