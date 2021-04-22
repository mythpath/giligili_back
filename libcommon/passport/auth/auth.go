package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/nerror"
	"selfText/giligili_back/libcommon/net/http/rest"
	"strings"
	"sync"
)

const (
	PassportModule   = "nerv-passport"
	PassportResource = "project"
)

type passportAccessList struct {
	passportHost string
}

type PassportClient struct {
	httpSender     *http.Client
	ac             *passportAccessList
	perm           map[string]*ModulePermission
	permMtx        *sync.Mutex
	localAddr      string
	tokenCache     map[string]string
	staticPermPath string

	Config brick.Config `inject:"config"`
}

func NewPassportClient(path string) *PassportClient {
	p := &PassportClient{}
	p.staticPermPath = path

	return p
}

func (p *PassportClient) Init() error {
	passportAddr := p.Config.GetMapString("auth", "passport_addr")
	localAddr := p.Config.GetMapString("auth", "local_addr")

	p.permMtx = &sync.Mutex{}
	p.perm = make(map[string]*ModulePermission)
	p.tokenCache = make(map[string]string)
	p.httpSender = &http.Client{}
	if !strings.HasPrefix(passportAddr, "http://") {
		passportAddr = "http://" + passportAddr
	}
	p.ac = &passportAccessList{
		passportHost: strings.TrimSuffix(passportAddr, "/"),
	}
	if !strings.HasPrefix(localAddr, "http://") {
		localAddr = "http://" + localAddr
	}
	p.localAddr = localAddr

	if len(p.staticPermPath) != 0 {
		data, err := ioutil.ReadFile(p.staticPermPath)
		if err != nil {
			panic(err.Error())
		}
		err = json.Unmarshal(data, &p.perm)
		if err != nil {
			panic(err.Error())
		}
		for module, _ := range p.perm {
			p.perm[module].LocalAddr = localAddr
		}
	}
	return nil
}

func (p *PassportClient) Register(
	module, mLabel, resource, rLabel, operation, oLabel string,
	isSystem bool) {
	p.permMtx.Lock()
	defer p.permMtx.Unlock()
	if _, ok := p.perm[module]; !ok {
		p.perm[module] = NewModulePermission(module, mLabel, p.localAddr)
	}
	p.perm[module].register(resource, rLabel, operation, oLabel, isSystem)
}

type RegisterArgs struct {
	Module, Resource, Operation string
	MLabel, RLabel, OLabel      string
	IsSystem                    bool
}

func (p *PassportClient) RegisterV2(arg *RegisterArgs) {
	p.permMtx.Lock()
	defer p.permMtx.Unlock()
	if _, ok := p.perm[arg.Module]; !ok {
		p.perm[arg.Module] = NewModulePermission(arg.Module, arg.MLabel, p.localAddr)
	}
	p.perm[arg.Module].register(arg.Resource, arg.RLabel, arg.Operation, arg.OLabel, arg.IsSystem)
}

func (p *PassportClient) GetPerm() map[string]*ModulePermission {
	return p.perm
}

type AuthTokenPair struct {
	Token      string `json:"token"`
	ModuleName string `json:"name"`
}

func (p *PassportClient) RegisterOperation() error {
	url := p.ac.passportHost + "/api/passport/objs/Authorization/RegisterAuthorization"
	for _, item := range p.perm {
		respBuf, err := p.send("POST", url, item, nil)
		if err != nil {
			return err
		}
		t := &AuthTokenPair{}
		err = json.Unmarshal(respBuf, t)
		if err != nil {
			return nerror.NewCommonError("%v", err)
		}
		if t.Token == "" {
			return nerror.NewCommonError("the `%s token is nil`", item.Name)
		}
		if t.ModuleName != item.Name {
			return nerror.NewCommonError("illegal module name, expect: %s, got: %s", item.Name, t.ModuleName)
		}
		p.tokenCache[t.ModuleName] = t.Token
	}

	return nil
}

func (p *PassportClient) GetAuthToken(module string) string {
	if module == PassportModule {
		// use any valid module token
		for _, t := range p.tokenCache {
			return t
		}
	}
	return p.tokenCache[module]
}

func (p *PassportClient) GetModuleToken(module string) string {
	return p.GetAuthToken(module)
}

type PassportAuthInfo struct {
	UserName   string                `json:"accountname"`
	IsAdmin    bool                  `json:"isAdmin"`
	TenantName string                `json:"tenant"`
	Projects   []PassportAuthProject `json:"projects"`
}

func (p *PassportAuthInfo) ContainsProject(project string, projectID uint) (*PassportAuthProject, bool) {
	for i, _ := range p.Projects {
		if (len(project) != 0 && p.Projects[i].Name == project) ||
			(projectID != 0 && p.Projects[i].ID == projectID) {
			return &p.Projects[i], true
		}
	}
	return nil, false
}

type PassportAuthProject struct {
	Name    string             `json:"projectname"`
	ID      uint               `json:"projectid"`
	Modules []ModulePermission `json:"modules"`
}

func (p *PassportAuthProject) ContainesModule(module string) (*ModulePermission, bool) {
	for i, _ := range p.Modules {
		if p.Modules[i].Name == module {
			return &p.Modules[i], true
		}
	}
	return nil, false
}

func (p *PassportClient) CheckAuthorization(user, project, module, resource,
	token, selfToken string, projectID uint) (*PassportAuthInfo, error) {
	url := p.ac.passportHost + "/api/passport/objs/Authorization/CheckAuthorization"
	jsonData := map[string]interface{}{
		"accountname":  user,
		"projectname":  project,
		"projectid":    projectID,
		"modulename":   module,
		"resourcetype": resource,
		"token":        token,
	}
	header := map[string]string{
		rest.GiligiliUser:  user,
		rest.GiligiliToken: selfToken,
	}
	respBuf, err := p.send("POST", url, jsonData, header)
	if err != nil {
		return nil, err
	}
	info := &PassportAuthInfo{}
	err = json.Unmarshal(respBuf, info)
	if err != nil {
		return nil, nerror.NewCommonError("%v", err)
	}
	return info, nil
}

func (p *PassportClient) isAdmin(info *PassportAuthInfo) bool {
	// passport server promises that nerv-stack use 'admin' to indentify superuser and
	// nerv-cloud use 'isAdmin' to indentify superuser
	if info.IsAdmin || info.UserName == "admin" {
		return true
	}
	return false
}

func (p *PassportClient) AuthByCtx(ctx context.Context, project, module, resource, operation string) error {
	user := ctx.Value(rest.CurrentUserKey()).(string)
	info := ctx.Value(rest.CurrentAuthInfoKey()).(*PassportAuthInfo)
	return p.AuthByInfo(info, user, project, module, resource, operation, 0)
}

func (p *PassportClient) AuthByCtxWithID(ctx context.Context, module, resource, operation string, projectID uint) error {
	user := ctx.Value(rest.CurrentUserKey()).(string)
	info := ctx.Value(rest.CurrentAuthInfoKey()).(*PassportAuthInfo)
	return p.AuthByInfo(info, user, "", module, resource, operation, projectID)
}

func (p *PassportClient) AuthByInfo(info *PassportAuthInfo, user, project, module, resource, operation string, projectID uint) error {
	if info == nil {
		return nerror.NewAuthorizationError("No passport auth info")
	}

	if p.isAdmin(info) {
		return nil
	}

	projectPerm, ok := info.ContainsProject(project, projectID)
	if !ok {
		return nerror.NewAuthorizationError("Has no permission on project [%s]", project)
	}
	modulePerm, ok := projectPerm.ContainesModule(module)
	if !ok {
		return nerror.NewAuthorizationError("Has no permission on module [%s]", module)
	}
	rscPerm, ok := modulePerm.Resource.Contains(resource)
	if !ok {
		return nerror.NewAuthorizationError("Has no permission on resource [%s]", resource)
	}
	ok = rscPerm.Operation.Contains(operation)
	if !ok {
		return nerror.NewAuthorizationError("Has no permission on operation [%s]", operation)
	}
	return nil
}

func (p *PassportClient) AuthWithoutOperationByCtx(ctx context.Context, project, module, resource string) error {
	user := ctx.Value(rest.CurrentUserKey()).(string)
	info := ctx.Value(rest.CurrentAuthInfoKey()).(*PassportAuthInfo)
	return p.AuthWithoutOperationByInfo(info, user, project, module, resource)
}

func (p *PassportClient) AuthWithoutOperationByInfo(info *PassportAuthInfo, user, project, module, resource string) error {
	if info == nil {
		return nerror.NewAuthorizationError("No passport auth info")
	}

	if p.isAdmin(info) {
		return nil
	}

	projectPerm, ok := info.ContainsProject(project, 0)
	if !ok {
		return nerror.NewAuthorizationError("Has no permission on project [%s]", project)
	}
	modulePerm, ok := projectPerm.ContainesModule(module)
	if !ok {
		return nerror.NewAuthorizationError("Has no permission on module [%s]", module)
	}
	_, ok = modulePerm.Resource.Contains(resource)
	if !ok {
		return nerror.NewAuthorizationError("Has no permission on resource [%s]", resource)
	}
	return nil
}

func (p *PassportClient) AuthAllByCtx(ctx context.Context, module, resource, operation string) (ProjectPairs, error) {
	user := ctx.Value(rest.CurrentUserKey()).(string)
	info := ctx.Value(rest.CurrentAuthInfoKey()).(*PassportAuthInfo)
	return p.AuthAllByInfo(info, user, module, resource, operation)
}

func (p *PassportClient) AuthAllByInfo(info *PassportAuthInfo, user, module, resource, operation string) (ProjectPairs, error) {
	if info == nil {
		return nil, nerror.NewAuthorizationError("No passport auth info")
	}

	if p.isAdmin(info) {
		result := NewProjectPairs()
		for i, _ := range info.Projects {
			result.Append(&NervProjectInfoOutL1{
				Name: info.Projects[i].Name,
				ID:   info.Projects[i].ID,
			})
		}
		return result, nil
	}

	result := NewProjectPairs()
	for i, _ := range info.Projects {
		modulePerm, ok := info.Projects[i].ContainesModule(module)
		if !ok {
			continue
		}
		rscPerm, ok := modulePerm.Resource.Contains(resource)
		if !ok {
			continue
		}
		ok = rscPerm.Operation.Contains(operation)
		if !ok {
			continue
		}
		result.Append(&NervProjectInfoOutL1{
			Name: info.Projects[i].Name,
			ID:   info.Projects[i].ID,
		})
	}
	if len(result) == 0 {
		return nil, nerror.NewAuthorizationError("No project has authorization")
	}
	return result, nil
}

type NervProjectInfoL0 struct {
	Project []NervProjectInfoL1 `json:"projects"`
}

type NervProjectInfoL1 struct {
	Name string `json:"name,omitempty"`
	ID   uint   `json:"id,omitempty"`
}

type NervProjectInfoOutL0 struct {
	Project []NervProjectInfoOutL1 `json:"projects"`
}

type NervProjectInfoOutL1 struct {
	Name string `json:"name"`
	ID   uint   `json:"id"`
}

type ProjectPairs []NervProjectInfoOutL1

func NewProjectPairs() ProjectPairs {
	p := ProjectPairs{}
	return p
}

func (p *ProjectPairs) Append(item *NervProjectInfoOutL1) {
	*p = append(*p, *item)
}

func (p *ProjectPairs) GetAllNervID() []uint {
	ret := make([]uint, len(*p))
	for i, _ := range *p {
		ret[i] = (*p)[i].ID
	}

	return ret
}

func (p *ProjectPairs) GetAllName() []string {
	ret := make([]string, len(*p))
	for i, _ := range *p {
		ret[i] = (*p)[i].Name
	}

	return ret
}

func (p *PassportClient) FindNervProjectID(names []string) (ProjectPairs, error) {
	url := p.ac.passportHost + "/api/passport/objs/getprojectidbyname"

	projects := &NervProjectInfoL0{}
	projects.Project = make([]NervProjectInfoL1, len(names))
	for i, _ := range projects.Project {
		projects.Project[i].Name = names[i]
	}

	respBuf, err := p.send("POST", url, projects, map[string]string{"NERV-USER": "admin"})
	if err != nil {
		return nil, err
	}
	out := &NervProjectInfoOutL0{}
	err = json.Unmarshal(respBuf, out)
	if err != nil {
		return nil, nerror.NewCommonError("%v", err)
	}

	pps := NewProjectPairs()
	for i, _ := range out.Project {
		if out.Project[i].ID != 0 {
			pps.Append(&out.Project[i])
		}
	}

	return pps, nil
}

func (p *PassportClient) FindNervProjectName(id []uint) (ProjectPairs, error) {
	url := p.ac.passportHost + "/api/passport/objs/getprojectnamebyid"

	projects := &NervProjectInfoL0{}
	projects.Project = make([]NervProjectInfoL1, len(id))
	for i, _ := range projects.Project {
		projects.Project[i].ID = id[i]
	}

	respBuf, err := p.send("POST", url, projects, map[string]string{"NERV-USER": "admin"})
	if err != nil {
		return nil, err
	}
	out := &NervProjectInfoOutL0{}
	err = json.Unmarshal(respBuf, out)
	if err != nil {
		return nil, nerror.NewCommonError("%v", err)
	}

	pps := NewProjectPairs()
	for i, _ := range out.Project {
		if out.Project[i].Name != "" {
			pps.Append(&out.Project[i])
		}
	}

	return pps, nil
}

func (p *PassportClient) send(method, url string, data interface{}, headers map[string]string) ([]byte, error) {
	var sendBody []byte
	var err error
	if data != nil {
		sendBody, err = json.Marshal(data)
		if err != nil {
			return nil, nerror.NewCommonError(err.Error())
		}
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(sendBody))
	if err != nil {
		return nil, nerror.NewCommonError(err.Error())
	}
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	resp, err := p.httpSender.Do(req)
	if err != nil {
		return nil, nerror.NewCommonError(err.Error())
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nerror.NewCommonError(err.Error())
	}
	if resp.StatusCode >= 400 {
		if resp.StatusCode == 403 {
			return nil, nerror.NewAuthorizationErrorForward(respBody)
		}
		return nil, nerror.NewCommonError("nerv-passport: %s", string(respBody))
	}
	return respBody, nil
}
