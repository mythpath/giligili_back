package file

import (
	"selfText/giligili_back/libcommon/net/http/rest/render"
	"io"
	"net/http"
	"os"

	"bytes"
	"crypto/md5"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
)

type UploadServer struct {
}

func (p *UploadServer) Handle(mx *chi.Mux, path string, fileRoot string) {
	if strings.ContainsAny(path, ":*") {
		panic("chi: FileServer does not permit URL parameters.")
	}
	upload := Upload(fileRoot)
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

type Upload string

func (p Upload) Post(w http.ResponseWriter, req *http.Request) {
	//fmt.Printf("upload: %s %s\n", req.URL.Path, p)
	req.ParseMultipartForm(32 << 20)
	file, handler, err := req.FormFile("uploadfile")
	if err != nil {
		render.Status(req, 500)
		render.JSON(w, req, err)
		return
	}
	defer file.Close()
	//fmt.Println(handler.Filename)
	path := filepath.Join(string(p), filepath.FromSlash(path.Clean("/"+req.URL.Path)))
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

	if err := WriteMD5Value(filepath.Join(path, handler.Filename), bytes.NewReader(hMd5.Sum(nil))); err != nil {
		render.Status(req, 500)
		render.JSON(w, req, err)
		return
	}
}
