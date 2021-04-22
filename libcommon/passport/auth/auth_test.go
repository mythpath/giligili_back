package auth

import (
	"selfText/giligili_back/libcommon/brick"
	"log"
	"sync"
	"testing"
)

// concurrent regist
func TestAuthRegisterResult(t *testing.T) {
	container := brick.NewContainer()
	container.Add(&brick.JSONConfigService{}, "config", brick.FactoryFunc(func() interface{} {
		return brick.NewJSONConfigService("config/test_config.json")
	}))
	container.Add(&PassportClient{}, "nerv-passport", nil)
	container.Build()
	testMain(container)
}

func TestPassportWithStaticPerm(t *testing.T) {
	container := brick.NewContainer()
	container.Add(&brick.JSONConfigService{}, "config", brick.FactoryFunc(func() interface{} {
		return brick.NewJSONConfigService("config/test_config.json")
	}))
	container.Add(&PassportClient{}, "nerv-passport", brick.FactoryFunc(func() interface{} {
		return NewPassportClient("config/static_permission.json")
	}))
	container.Build()
	passport := container.GetByName("nerv-passport").(*PassportClient)
	err := passport.RegisterOperation()
	if err != nil {
		t.Errorf("error: %s", err.Error())
	}
	t.Logf("get token: %s", passport.GetAuthToken("test-module"))
}

func testMain(container *brick.Container) {
	passport := container.GetByName("nerv-passport").(*PassportClient)

	funcWrap := func(wg *sync.WaitGroup,
		module string, fn func(wg *sync.WaitGroup, module string)) func() {
		return func() {
			go func(module string) {
				fn(wg, module)
				wg.Done()
			}(module)
		}
	}
	wg := sync.WaitGroup{}
	func1 := func(wg *sync.WaitGroup, module string) {
		log.Printf("func1")
		passport.Register(module, "测试模块", "1", "一", "action1", "操作1", false)
	}
	func2 := func(wg *sync.WaitGroup, module string) {
		log.Printf("func2")
		passport.Register(module, "测试模块", "2", "二", "action1", "操作1", false)
	}
	func3 := func(wg *sync.WaitGroup, module string) {
		log.Printf("func3")
		passport.Register(module, "测试模块", "1", "一", "action2", "操作2", false)
	}
	func4 := func(wg *sync.WaitGroup, module string) {
		log.Printf("func4")
		passport.Register(module, "测试模块", "1", "一", "action3", "操作3", true)
	}
	func5 := func(wg *sync.WaitGroup, module string) {
		log.Printf("func5")
		passport.Register(module, "测试模块", "1", "一", "action3", "操作3", true)
	}
	func6 := func(wg *sync.WaitGroup, module string) {
		log.Printf("func5")
		passport.Register(module, "测试模块", "3", "三", "", "", true)
	}

	modelList := []string{
		"test-module", "test-module1", "test-module2", "test-module3", "test-module4",
	}
	funcList := make(map[string][]func())
	for _, module := range modelList {
		funcList[module] = []func(){
			funcWrap(&wg, module, func1),
			funcWrap(&wg, module, func2),
			funcWrap(&wg, module, func3),
			funcWrap(&wg, module, func4),
			funcWrap(&wg, module, func5),
			funcWrap(&wg, module, func6),
		}
	}

	for _, flist := range funcList {
		for i := 0; i < len(funcList); i++ {
			wg.Add(1)
			flist[i]()
		}
	}
	wg.Wait()

	//tmp, _ := json.MarshalIndent(&passport.perm, "", "  ")
	//log.Printf("appPerm: %s", string(tmp))

	err := passport.RegisterOperation()
	if err != nil {
		log.Printf("error: %s", err.Error())
	}
	for _, module := range modelList {
		log.Printf("get token: %s", passport.GetAuthToken(module))
	}
}

func TestCheckAuthorization(t *testing.T) {
	container := brick.NewContainer()
	container.Add(&brick.JSONConfigService{}, "config", brick.FactoryFunc(func() interface{} {
		return brick.NewJSONConfigService("config/test_config.json")
	}))
	container.Add(&PassportClient{}, "nerv-passport", brick.FactoryFunc(func() interface{} {
		return NewPassportClient("config/static_permission.json")
	}))
	container.Build()
	passport := container.GetByName("nerv-passport").(*PassportClient)
	err := passport.RegisterOperation()
	if err != nil {
		t.Errorf("error: %s", err.Error())
	}
	t.Logf("get token: %s", passport.GetAuthToken("test-module"))
	info, err := passport.CheckAuthorization("liuqing_123", "test", "test-module", "1", "a66dd3be0f17b246106ee5b9638da797",
		passport.GetAuthToken("test-module"), 0)
	if err != nil {
		t.Errorf("error: %s", err.Error())
	}
	t.Logf("info: %#v", *info)
}
