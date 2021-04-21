package service

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"net/http"
	"net/http/pprof"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/metrics"
	"selfText/giligili_back/libcommon/net/http/invoker"
	"selfText/giligili_back/libcommon/net/http/rest/render"
)

var (
	ApiObjsFilm      = "/api/objs/film"
	ApiObjsPerformer = "/api/objs/performer"
	ApiObjsUser      = "/api/objs/user"
	ApiObjsVideo     = "/api/objs/video"
	ApiObjsRank      = "/api/objs/rank"
	ApiObjsUpload    = "/api/objs/Upload"
	ApiHealth        = "/api/health"
)

type Http struct {
	Config    brick.Config                 `inject:"config"`
	Logger    *logging.LoggerService       `inject:"LoggerService"`
	Invoker   *invoker.Invoker             `inject:"Invoker"`
	MExporter *metrics.HttpExporterService `inject:"MetricsHttpExporter"`
}

func (h *Http) Init() error {
	r := chi.NewRouter()
	// set request log
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger:  h.Logger.Logger,
		NoColor: false,
	}))

	// pprof
	h.pprofRouter(r)

	h.filmRouter(r)
	h.videoRouter(r)
	h.performerRouter(r)
	h.userRouter(r)
	h.uploadRouter(r)
	h.rankRouter(r)

	h.healthRouter(r)

	port := h.Config.GetMapInt("http", "port", 3353)
	go func() {
		h.Logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
	}()

	return nil
}

func (h *Http) Dispose() error {
	return nil
}

func (h *Http) healthRouter(r chi.Router) {
	r.Route(ApiHealth, func(r chi.Router) {
		r.Get("/", h.healthHandler)
	})
}

func (h *Http) healthHandler(w http.ResponseWriter, req *http.Request) {
	render.Status(req, http.StatusOK)
	render.TextJSON(w, req, "ok")
}

func (h *Http) pprofRouter(r chi.Router) {
	r.Route("/debug", func(r chi.Router) {
		r.Get("/pprof/profile", pprof.Profile)
		r.Get("/pprof/cmdline", pprof.Cmdline)
		r.Get("/pprof/symbol", pprof.Symbol)
		r.Get("/pprof/trace", pprof.Trace)
		r.Get("/pprof/*", pprof.Index)
	})
}

func (h *Http) performerRouter(r chi.Router) {
	handler := func(class, method string) http.HandlerFunc {
		return h.Invoker.Execute(class, method)
	}

	r.Route(ApiObjsPerformer, func(r chi.Router) {
		r.Post("/", handler("PerformerService", "Create"))
		r.Put("/", handler("PerformerService", "Update"))
		r.Get("/", handler("PerformerService", "List"))
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handler("PerformerService", "Get"))
			r.Delete("/", handler("PerformerService", "Delete"))
		})
	})
}

func (h *Http) filmRouter(r chi.Router) {
	handler := func(class, method string) http.HandlerFunc {
		return h.Invoker.Execute(class, method)
	}

	r.Route(ApiObjsFilm, func(r chi.Router) {
		r.Post("/", handler("FilmService", "Create"))
		r.Put("/", handler("FilmService", "Update"))
		r.Get("/", handler("FilmService", "List"))
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handler("FilmService", "Get"))
			r.Delete("/", handler("FilmService", "Delete"))
		})
	})
}

func (h *Http) userRouter(r chi.Router) {
	handler := func(class, method string) http.HandlerFunc {
		return h.Invoker.Execute(class, method)
	}

	r.Route(ApiObjsUser, func(r chi.Router) {
		r.Post("/", handler("UserService", "Create"))
		r.Put("/", handler("UserService", "Update"))
		r.Get("/", handler("UserService", "List"))
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handler("UserService", "Get"))
			r.Delete("/", handler("UserService", "Delete"))
		})

		r.Get("/me", handler("UserService", "GetMe"))
		r.Post("/register", handler("UserService", "Register"))
		r.Post("/login", handler("UserService", "Login"))
		r.Delete("/logout", handler("UserService", "Logout"))
	})
}

func (h *Http) videoRouter(r chi.Router) {
	handler := func(class, method string) http.HandlerFunc {
		return h.Invoker.Execute(class, method)
	}

	r.Route(ApiObjsVideo, func(r chi.Router) {
		r.Post("/", handler("VideoService", "Create"))
		r.Put("/", handler("VideoService", "Update"))
		r.Get("/", handler("VideoService", "List"))
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handler("VideoService", "Get"))
			r.Delete("/", handler("VideoService", "Delete"))
		})
	})
}

func (h *Http) rankRouter(r chi.Router) {
	handler := func(class, method string) http.HandlerFunc {
		return h.Invoker.Execute(class, method)
	}

	r.Route(ApiObjsRank, func(r chi.Router) {
		r.Get("/daily", handler("RankService", "DailyRank"))
	})
}

func (h *Http) uploadRouter(r chi.Router) {
	handler := func(class, method string) http.HandlerFunc {
		return h.Invoker.Execute(class, method)
	}

	r.Route(ApiObjsUpload, func(r chi.Router) {
		r.Post("/token", handler("UploadService", "UploadToken"))
	})
}
