package file

import (
	"net/http"
	"selfText/giligili_back/service/file/common"
	"strings"

	"github.com/go-chi/chi"
)

// FileRouting
type FileRouting struct {
}

func (p *FileRouting) Handle(mx *chi.Mux, path string, root FileSystem, lockBucket *common.ResourceLockBucket) {
	if strings.ContainsAny(path, ":*") {
		panic("chi: FileServer does not permit URL parameters.")
	}

	fs := FileServer(root, lockBucket)
	prefix := path
	path += "*"
	mx.Get(path, p.exec(prefix, func(w http.ResponseWriter, r *http.Request) {
		fs.Get(w, r)
	}))

	mx.Post(path, p.exec(prefix, func(w http.ResponseWriter, r *http.Request) {
		fs.Post(w, r)
	}))
	mx.Put(path, p.exec(prefix, func(w http.ResponseWriter, r *http.Request) {
		fs.Put(w, r)
	}))
	mx.Delete(path, p.exec(prefix, func(w http.ResponseWriter, r *http.Request) {
		fs.Delete(w, r)
	}))
}

func (p *FileRouting) exec(prefix string, fn http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
			r.URL.Path = p
			fn(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
}
