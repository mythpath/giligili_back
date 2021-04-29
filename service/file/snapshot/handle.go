package snapshot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"selfText/giligili_back/libcommon/net/http/rest/render"
	"selfText/giligili_back/service/file/common"
	"selfText/giligili_back/service/file/file"
	"strings"

	"github.com/sirupsen/logrus"
)

type SnapshotHandle struct {
	prefixUrl string          // /api/snapshots
	root      file.FileSystem // ${NERV_FILE_HOME}/snapshots
	metaH     *MetaHandle
	wsH       *WorkspaceHandle
	url       *Url

	wshFactory *WorkspaceFactory // develop workspace handle
	lockBucket *common.ResourceLockBucket
}

type result struct {
	name string
	err  error
}

func NewSnapshotsHandle(wshFactory *WorkspaceFactory, prefixUrl string,
	root file.FileSystem, metaHandle *MetaHandle, handle *WorkspaceHandle,
	url *Url, lockBucket *common.ResourceLockBucket) *SnapshotHandle {
	return &SnapshotHandle{
		wshFactory: wshFactory,
		prefixUrl:  prefixUrl,
		root:       root,
		metaH:      metaHandle,
		wsH:        handle,
		url:        url,
		lockBucket: lockBucket,
	}
}

// r.URL.Path - /nerv-app/2/api/scripts/nerv-app/type.json
func (p *SnapshotHandle) Get(w http.ResponseWriter, r *http.Request) {

	if !p.isExistSnapshot(p.url.Snapshot(), p.root) {
		render.Status(r, http.StatusBadRequest)
		render.TextJSON(w, r, fmt.Errorf("error snapshot: %s", p.url.Snapshot()))
		return
	}

	dir, ok := p.root.(file.Dir)
	if !ok {
		dir = file.Dir("")
	}
	rscLock := p.lockBucket.NewResourceLockV2(string(dir), p.url.Path())
	defer p.lockBucket.ReleaseLock(rscLock)
	rscLock.Lock()
	defer rscLock.Unlock()
	dMetaS, err := p.metaH.ReadM(p.url.Path())
	if err != nil && os.IsNotExist(err) {
		meta, err := p.createSnapshot()
		if err != nil {
			logrus.Errorf("failed to create snapshot: %v", err)
			render.Status(r, http.StatusInternalServerError)
			render.TextJSON(w, r, err)
			return
		}
		p.serverContent(w, r, meta)
		return
	}

	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.TextJSON(w, r, fmt.Errorf("error read meta data"))
		logrus.WithField("path", p.url.Path()).Errorf("failed to read meta data in path: %v", err)
		return
	}

	locked, err := p.metaH.Locked(p.url.MetaPath())
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.TextJSON(w, r, fmt.Errorf("error stat meta lock, err:%v", err))
		logrus.WithField("path", p.url.MetaPath()).Errorf("failed stat meta lock file in path: %v", err)
		return
	}

	if !locked {
		// create snapshot and server content
		dMetaS, err = p.createSnapshotByMeta(dMetaS)
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.TextJSON(w, r, err)
			return
		}
	}

	p.serverContent(w, r, dMetaS)
}

func (p *SnapshotHandle) Post(w http.ResponseWriter, r *http.Request) {
	route := &Route{}

	if err := render.Bind(r.Body, route); err != nil {
		render.Status(r, 400)
		render.TextJSON(w, r, fmt.Sprintf("parse json error: %s", err.Error()))
		return
	}

	if route.DeployId == 0 {
		render.Status(r, http.StatusBadRequest)
		render.TextJSON(w, r, "invalid params as DeployId == 0")
		return
	}
	route.Url = p.prefixUrl

	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath += "/"
		r.URL.Path = upath
	}

	dir, ok := p.root.(file.Dir)
	if !ok {
		dir = file.Dir("")
	}
	rscLock := p.lockBucket.NewResourceLockV2(string(dir), r.URL.Path)
	defer p.lockBucket.ReleaseLock(rscLock)
	rscLock.Lock()
	defer rscLock.Unlock()

	if route.BaseId != 0 {
		name := createUrl(filepath.Join(r.URL.Path, fmt.Sprintf("%d", route.BaseId)))
		if !p.isExistSnapshot(name, p.root) {
			logrus.Errorf("the base id:%d is not exist, deployId:%d", route.BaseId, route.DeployId)
		}
	}

	if err := p.createSnapshots(path.Clean(upath), route); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.TextJSON(w, r, err)
		return
	}

	b, err := json.Marshal(route)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.TextJSON(w, r, err)
		return
	}
	w.Write(b)
}

// lock
func (p *SnapshotHandle) Patch(w http.ResponseWriter, r *http.Request) {
	route, err := p.parse(r.Body)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.TextJSON(w, r, route)
		return
	}

	metaDir := filepath.Join(r.URL.Path, fmt.Sprintf("%d", route.DeployId))
	name := createUrl(metaDir)
	dir, ok := p.root.(file.Dir)
	if !ok {
		dir = file.Dir("")
	}
	rscLock := p.lockBucket.NewResourceLockV2(string(dir), r.URL.Path)
	defer p.lockBucket.ReleaseLock(rscLock)
	rscLock.Lock()
	defer rscLock.Unlock()

	if !p.isExistSnapshot(name, p.root) {
		render.Status(r, http.StatusBadRequest)
		render.TextJSON(w, r, fmt.Errorf("invalid snapshot: %s", name))
		return
	}

	lock, err := p.metaH.Locked(metaDir)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.TextJSON(w, r, fmt.Errorf("check locked because %v", err))
		return
	}

	if !lock {
		if err := p.metaH.Lock(metaDir); err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.TextJSON(w, r, fmt.Errorf("failed lock because %v", err))
			return
		}
	}
}

func (p *SnapshotHandle) Recover(w http.ResponseWriter, r *http.Request) {
	route := &Route{}

	if err := render.Bind(r.Body, route); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.TextJSON(w, r, err)
		return
	}

	if route.DeployId == 0 {
		render.Status(r, http.StatusBadRequest)
		render.TextJSON(w, r, "invalid params")
		return
	}

	metaName := filepath.Join(r.URL.Path, fmt.Sprintf("%d", route.DeployId))
	name := createUrl(metaName)
	if !p.isExistSnapshot(name, p.root) {
		render.Status(r, http.StatusBadRequest)
		render.TextJSON(w, r, fmt.Errorf("invalid snapshot: %s", name))
		return
	}

	if err := p.rollback(metaName, route.DeployId); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.TextJSON(w, r, fmt.Errorf("error rollback snapshot: %s", metaName))
		logrus.WithField("name", name).Errorf("failed to rollback snapshot: %v", err)
		return
	}

	render.Status(r, http.StatusOK)
	render.TextJSON(w, r, "")
}

func (p *SnapshotHandle) Delete(w http.ResponseWriter, r *http.Request) {
	resourceLocation, err := NewUrl(r.URL.Path)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.TextJSON(w, r, fmt.Errorf("invalid url: %s", err.Error()))
		return
	}
	err = p.root.Delete(resourceLocation.Snapshot())
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.TextJSON(w, r, fmt.Errorf("err: %s", err.Error()))
		return
	}
	err = p.metaH.root.Delete(resourceLocation.MetaPath())
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.TextJSON(w, r, fmt.Errorf("err: %s", err.Error()))
		return
	}
	logrus.Printf("Deleted resource location: %#v", *resourceLocation)

	render.Status(r, http.StatusOK)
	render.TextJSON(w, r, "")
}

// path --> /nerv-app/2
func (p *SnapshotHandle) rollback(path string, deployId uint) error {
	upath := path
	if !strings.HasPrefix(upath, "/") {
		upath = filepath.Join("/", upath)
	}
	sepCount := strings.Count(upath, "/")
	filenames, err := p.metaH.ReadFiles(upath)
	if err != nil {
		logrus.WithField("path", path).Errorf("failed to read path: %v", err)
		return err
	}

	recoverF := func(filename string, deployId uint) <-chan result {
		rs := make(chan result, 1)

		go func() {
			defer close(rs)

			// ---> /3/nerv-app/2/api/scripts/nerv-app/type.json
			url, err := NewUrl(filepath.Join(fmt.Sprintf("/%d", sepCount+1), filename))
			if err != nil {
				rs <- result{name: filename, err: err}

				return
			}

			// ---> /api/scripts/nerv-app/type.json
			workspaceH, err := p.wshFactory.NewWorkspaceHandle(url.RealPath())
			if err != nil {
				rs <- result{name: filename, err: err}

				return
			}

			// ---> /nerv-app/2/api/scripts/nerv-app/type.json
			meta, err := p.metaH.ReadM(url.Path())
			if err != nil {
				rs <- result{name: filename, err: err}

				return
			}

			// only rollback file
			fileMeta, ok := meta.(*MetaSnapshot)
			if !ok {
				return
			}

			err = p.rollbackByMeta(workspaceH, url, fileMeta, deployId)
			if err != nil {
				rs <- result{name: filename, err: err}

				return
			}

			rs <- result{name: filename, err: nil}
		}()

		return rs
	}

	results := make([]<-chan result, 0, len(filenames))
	for _, filename := range filenames {
		results = append(results, recoverF(filename, deployId))
	}

	for _, chanR := range results {
		select {
		case rs := <-chanR:
			if rs.err != nil {
				if err == nil {
					err = rs.err
				}

				logrus.Errorf("failed to recover %s because %v", rs.name, rs.err)
			}
		}
	}

	return err
}

func (p *SnapshotHandle) rollbackByMeta(wsh *WorkspaceHandle, url *Url, meta *MetaSnapshot, deployId uint) error {
	filename := p.metaH.RefFilename(url.Path(), meta)
	f, err := p.root.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	err = wsh.CopyFileFromSnapshots(url, f)
	if err != nil {
		return err
	}

	return wsh.WriteMD5(url, meta.MD5)
}

func (p *SnapshotHandle) createSnapshotByMeta(meta Meta) (Meta, error) {
	switch vt := meta.(type) {
	case *DirMetaSnapshot:
		return p.createDirSnapshotByMeta(vt)
	case *MetaSnapshot:
		return p.createFileSnapshotByMeta(vt)
	}

	logrus.Errorf("invalid meta type: %s", meta.Type())
	return meta, errors.New("invalid meta type")
}

func (p *SnapshotHandle) createFileSnapshotByMeta(meta *MetaSnapshot) (*MetaSnapshot, error) {
	md5, err := p.wsH.Md5(p.url)
	if err != nil {
		logrus.WithField("path", p.url.Path()).Errorf("failed to read md5 content in path: %v", err)
		return nil, fmt.Errorf("error read md5 content")
	}

	nMetaS := meta.Clone()
	if nMetaS.MD5 != md5 {
		// copy
		nMetaS.MD5 = md5
		nMetaS.RefUrl = p.url.UrlPath()

		err = p.copyFileFromWorkspace(p.url, p.wsH)
		if err != nil {
			logrus.WithField("url", p.url.UrlPath()).Errorf("failed to copy workspace to snapshot: %v", err)
			return nil, fmt.Errorf("error copy workspace, %v", err)
		}

		err = p.metaH.WriteM(p.url.Path(), nMetaS)
		if err != nil {
			logrus.Errorf("failed to write meta data<%v> to %s: %v", nMetaS, p.url.Path(), err)
			return nil, fmt.Errorf("error write meta data")
		}
	}

	return nMetaS, nil
}

func (p *SnapshotHandle) createDirSnapshotByMeta(meta *DirMetaSnapshot) (Meta, error) {
	wf, err := p.wsH.Open(p.url)
	if err != nil {
		logrus.WithField("url", p.url.Path()).Errorf("failed to open workspace: %v", err)
		return nil, fmt.Errorf("failed to open url: %s", p.url.Path())
	}
	defer wf.Close()

	wfi, err := wf.Stat()
	if err != nil {
		logrus.WithField("url", p.url.Path()).Errorf("failed to stat workspace: %v", err)
		return nil, fmt.Errorf("failed to stat url: %s", p.url.Path())
	}

	// directory -> file
	if !wfi.IsDir() {
		//logrus.WithField("url", p.url.Path()).Warn("the path is not directory")
		// remove meta data
		if err := p.metaH.RemoveMeta(p.url.Path(), meta); err != nil {
			return nil, err
		}

		// recreate snapshot
		meta, err := p.createFileSnapshot()
		if err != nil {
			logrus.WithField("url", p.url.Path()).Errorf("failed to create file snapshot: %v", err)
			return nil, err
		}

		return meta, nil
	}

	wfInfos, err := wf.Readdir(-1)
	if err != nil {
		logrus.WithField("url", p.url.Path()).Errorf("failed to read directory: %v", err)
		return nil, err
	}
	wfiles, _ := EncodeDirs(wfInfos, filepath.Join("/", p.url.SubPath(p.wsH.prefix)))
	dMeta := &DirMetaSnapshot{
		Url:   p.url.Path(),
		Files: wfiles,
	}
	if err := p.metaH.WriteM(p.url.Path(), dMeta); err != nil {
		logrus.WithField("url", p.url.Path()).Errorf("failed to write meta data: %v", err)
		return nil, err
	}

	return dMeta, nil
}

func (p *SnapshotHandle) createSnapshot() (Meta, error) {
	return p.createDirSnapshotByMeta(nil)
}

func (p *SnapshotHandle) createFileSnapshot() (Meta, error) {
	md5, err := p.wsH.Md5(p.url)
	if err != nil {
		logrus.WithField("url", p.url.Path()).Errorf("failed to read md5 data: %v", err)
		return nil, fmt.Errorf("error read md5 data")
	}

	err = p.copyFileFromWorkspace(p.url, p.wsH)
	if err != nil {
		logrus.WithField("url", p.url.Path()).Errorf("failed to copy content from workspace: %v", err)
		return nil, err
	}

	metaS := &MetaSnapshot{
		MD5:         md5,
		RefUrl:      p.url.UrlPath(),
		Description: make(map[string]string),
	}

	err = p.metaH.WriteM(p.url.Path(), metaS)
	if err != nil {
		logrus.WithField("url", p.url.Path()).Errorf("failed to write meta data<%s>: %v", metaS, err)
		return nil, fmt.Errorf("error write meta data")
	}

	return metaS, nil
}

func (p *SnapshotHandle) copy(w http.ResponseWriter, r *http.Request) {
	md5, err := p.wsH.Md5(p.url)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.TextJSON(w, r, fmt.Errorf("error read md5 data"))
		logrus.Errorf("Failed to read %s md5 data, err:%v", p.url.Path(), err)
		return
	}

	err = p.copyFileFromWorkspace(p.url, p.wsH)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.TextJSON(w, r, fmt.Errorf("error copy workspace, %v", err))
		return
	}

	metaS := &MetaSnapshot{
		MD5:         md5,
		RefUrl:      p.url.UrlPath(),
		Description: make(map[string]string),
	}

	err = p.metaH.WriteM(p.url.Path(), metaS)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.TextJSON(w, r, fmt.Errorf("error write meta data"))
		logrus.Errorf("Failed to write %s meta data<%v>, err:%v", p.url.Path(), metaS, err)
		return
	}

	p.serverContent(w, r, metaS)
}

func (p *SnapshotHandle) serverContent(w http.ResponseWriter, r *http.Request, meta Meta) {
	switch vt := meta.(type) {
	case *MetaSnapshot:
		realPath := p.metaH.RefFilename(p.url.Path(), vt)
		r.URL.Path = filepath.Clean(realPath)
		p.serverSnapshots(w, r)
	case *DirMetaSnapshot:
		r.URL.Path = vt.Url
		bs, err := json.Marshal(vt.Files)
		if err != nil {
			logrus.Errorf("failed marshal files: %v", vt.Files)
			render.Status(r, http.StatusInternalServerError)
			render.TextJSON(w, r, fmt.Errorf("error encode files"))
			return
		}

		render.Status(r, http.StatusOK)
		render.TextJSON(w, r, string(bs))
	default:
		render.Status(r, http.StatusBadRequest)
		render.TextJSON(w, r, fmt.Errorf("bad request"))
	}
}

func (p *SnapshotHandle) copyFileFromWorkspace(url *Url, wsh *WorkspaceHandle) error {
	dir, ok := p.root.(file.Dir)
	if !ok {
		return fmt.Errorf("invalid dir %v", p.root)
	}

	filename := filepath.Join(string(dir), url.UrlPath())
	dirPath := filepath.Dir(filename)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		logrus.Errorf("Failed to mkdir %s, err:%v", dirPath, err)
		return err
	}

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	wshF, err := wsh.Open(url)
	if err != nil {
		return err
	}
	defer wsh.Close(wshF)

	_, err = io.Copy(f, wshF)

	return nil
}

func (p *SnapshotHandle) isExistSnapshot(filename string, fs file.FileSystem) bool {
	_, err := fs.Stat(filename)
	if err != nil {
		return false
	}

	return true
}

func (p *SnapshotHandle) createSnapshots(path string, route *Route) error {
	dirName := filepath.Join("/", path, fmt.Sprintf("%d", route.DeployId))
	url := createUrl(dirName)
	if err := p.root.MkdirAll(url, os.ModePerm); err != nil {
		return err
	}
	route.Url = filepath.Join(route.Url, url)

	locked, err := p.metaH.Locked(dirName)
	if err != nil {
		logrus.WithField("dirname", dirName).Errorf("failed to judge exist locked: %v", err)
		return err
	}

	if locked {
		return nil
	}

	if route.BaseId != 0 {
		srcName := filepath.Join(path, fmt.Sprintf("%d", route.BaseId))

		if route.Copy {
			return p.metaH.CopyContentMeta(dirName, srcName)
		}

		return p.metaH.HardLinkMeta(dirName, srcName)
	}

	return nil
}

func (p *SnapshotHandle) parse(r io.ReadCloser) (*Route, error) {
	route := &Route{}

	if err := render.Bind(r, route); err != nil {
		return nil, err
	}

	if route.DeployId == 0 {
		return nil, fmt.Errorf("invalid param")
	}

	return route, nil
}

func createUrl(dirname string) string {
	if !strings.HasPrefix(dirname, "/") {
		dirname += "/"
	}

	// /3/nerv-app/2
	// /2/2
	sepCount := strings.Count(dirname, "/")

	return filepath.Join(fmt.Sprintf("/%d/%s", sepCount+1, dirname))
}

func (p *SnapshotHandle) serverSnapshots(w http.ResponseWriter, r *http.Request) {
	f := file.FileServer(p.root, p.lockBucket)
	f.Get(w, r)
}
