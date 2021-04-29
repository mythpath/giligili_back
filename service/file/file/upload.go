package file

import (
	"io"
	"net/http"
	"os"
	"selfText/giligili_back/libcommon/net/http/rest/render"
	"selfText/giligili_back/service/file/common"

	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"

	"crypto/md5"
	"path"
	"path/filepath"
	"strings"
)

type UploadServer struct {
}

func (p *UploadServer) Handle(mx *chi.Mux, path string, fileRoot string, lockBucket *common.ResourceLockBucket) {
	if strings.ContainsAny(path, ":*") {
		panic("chi: FileServer does not permit URL parameters.")
	}
	upload := NewUpload(fileRoot, lockBucket)
	prefix := path
	path += "*"

	mx.Post(path, p.exec(prefix, func(w http.ResponseWriter, r *http.Request) {
		upload.Post(w, r)
	}))
}

func (p *UploadServer) exec(prefix string, fn http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
			r.URL.Path = p
			fn(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
}

type Upload struct {
	fileRoot   string
	lockBucket *common.ResourceLockBucket
}

func NewUpload(fileRoot string, lockBucket *common.ResourceLockBucket) *Upload {
	return &Upload{fileRoot, lockBucket}
}

func (p Upload) Post(w http.ResponseWriter, req *http.Request) {
	rscLock := p.lockBucket.NewResourceLockV2(string(p.fileRoot), req.URL.Path)
	defer p.lockBucket.ReleaseLock(rscLock)
	rscLock.Lock()
	defer rscLock.Unlock()

	nervName := req.Header.Get("NERV-USER")
	err := req.ParseMultipartForm(32 << 20)
	if err != nil {
		render.Status(req, 500)
		render.JSON(w, req, err)
		logrus.Errorf("parse form error: %+v", err)
		req.MultipartForm.RemoveAll()
		return
	}
	file, handler, err := req.FormFile("uploadfile")
	if err != nil {
		render.Status(req, 500)
		render.JSON(w, req, err)
		logrus.Errorf("get form key error: %+v", err)
		return
	}
	defer file.Close()
	defer req.MultipartForm.RemoveAll()
	path := filepath.Join(p.fileRoot, filepath.FromSlash(path.Clean("/"+req.URL.Path)))
	err = os.MkdirAll(path, 0777)
	if err != nil {
		render.Status(req, 500)
		render.JSON(w, req, err)
		return
	}

	f, err := os.OpenFile(path+"/"+handler.Filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		render.Status(req, 500)
		render.JSON(w, req, err)
		return
	}
	defer f.Close()

	hMd5 := md5.New()
	writer := io.MultiWriter(hMd5, f)

	_, err = io.Copy(writer, file)
	if err != nil {
		render.Status(req, 500)
		render.JSON(w, req, err)
		return
	}

	_, err = createMeta(Dir(p.fileRoot), filepath.Join("/"+req.URL.Path, handler.Filename), "file", "", nervName)
	if err != nil {
		render.Status(req, 500)
		render.JSON(w, req, err)
		return
	}

	if err := WriteMD5Value(filepath.Join(path, handler.Filename), hMd5); err != nil {
		render.Status(req, 500)
		render.JSON(w, req, err)
		return
	}
}
