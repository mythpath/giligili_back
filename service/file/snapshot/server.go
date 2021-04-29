package snapshot

import (
	"fmt"
	"net/http"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/net/http/rest/render"
	"selfText/giligili_back/service/file/common"
	"selfText/giligili_back/service/file/file"
	"strings"

	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
)

var (
	snapshotRoute     = "/api/snapshots"
	metaSnapshotRoute = "/api/meta_snapshots"
)

type SnapshotServer struct {
}

func (p *SnapshotServer) Init(r *chi.Mux, config brick.Config,
	templateF func(string) string, lockBucket *common.ResourceLockBucket) error {
	sv := templateF(config.GetMapString("snapshots", snapshotRoute))
	msv := templateF(config.GetMapString("snapshots", metaSnapshotRoute))

	r.Route(snapshotRoute, func(r chi.Router) {
		metaSnapshotHandle := NewMetaSnapshotHandle(file.Dir(msv))
		wshFactory := NewWorkspaceFactory(config, templateF)
		handle := NewSnapshotsHandle(wshFactory, snapshotRoute, file.Dir(sv),
			metaSnapshotHandle, nil, nil, lockBucket)

		r.Post("/*", p.exec(snapshotRoute, func(w http.ResponseWriter, r *http.Request) {
			handle.Post(w, r)
		}))

		r.Put("/*", p.exec(snapshotRoute, func(w http.ResponseWriter, r *http.Request) {
			handle.Recover(w, r)
		}))

		r.Patch("/*", p.exec(snapshotRoute, func(w http.ResponseWriter, r *http.Request) {
			handle.Patch(w, r)
		}))

		r.Get("/*", p.exec(snapshotRoute, func(w http.ResponseWriter, r *http.Request) {
			handle := NewSnapshotsHandle(wshFactory, snapshotRoute, file.Dir(sv),
				metaSnapshotHandle, nil, nil, lockBucket)
			l, err := NewUrl(r.URL.Path)
			if err != nil {
				render.Status(r, http.StatusBadRequest)
				render.TextJSON(w, r, fmt.Sprintf("error url: %s, err: %s", r.URL.Path, err.Error()))
				logrus.Infof("failed to get real url, url:%s, err:%v", r.URL.Path, err)
				return
			}
			handle.url = l

			workspaceH, err := wshFactory.NewWorkspaceHandle(l.path)
			if err != nil {
				render.Status(r, http.StatusBadRequest)
				render.TextJSON(w, r, fmt.Sprintf("error url: %s, err: %s", r.URL.Path, err.Error()))
				logrus.Infof("failed to get workspace handle, url:%s, err:%v", r.URL.Path, err)
				return
			}

			handle.wsH = workspaceH
			handle.Get(w, r)
		}))

		r.Delete("/*", p.exec(snapshotRoute, func(w http.ResponseWriter, r *http.Request) {
			handle.Delete(w, r)
		}))

		logrus.Infof("snapshot router: %s -> %s", snapshotRoute, sv)
	})

	// /api/meta_snapshots
	r.Route(metaSnapshotRoute, func(r chi.Router) {
		fs := file.FileServer(file.Dir(msv), lockBucket)

		r.Get("/*", p.exec(metaSnapshotRoute, func(w http.ResponseWriter, r *http.Request) {
			fs.Get(w, r)
		}))

		logrus.Infof("meta snapshot router: %s -> %s", metaSnapshotRoute, msv)
	})

	return nil
}

func (p *SnapshotServer) exec(prefix string, fn http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
			if !strings.HasPrefix(p, "/") {
				p = "/" + p
			}
			r.URL.Path = p
			fn(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
}
