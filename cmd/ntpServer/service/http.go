package service

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/net/http/invoker"
	"time"
)

var (
	ApiObjsTime = "/api/objs/time"
)

type Http struct {
	Config  brick.Config           `inject:"config"`
	Logger  *logging.LoggerService `inject:"LoggerService"`
	Invoker *invoker.Invoker       `inject:"Invoker"`
}

func (h *Http) Init() error {
	r := chi.NewRouter()
	// set request log
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger:  h.Logger.Logger,
		NoColor: false,
	}))

	port := h.Config.GetMapInt("http", "port", 3232)
	svcFlag := h.Config.GetMapString("http", "type", "in")

	if svcFlag == "in" {
		// pprof
		h.pprofRouter(r)

		h.remoteTimeRouter(r)

		go func() {
			h.Logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
		}()
	} else {
		addr, err := net.ResolveTCPAddr("tcp", string(rune(port)))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
			os.Exit(1)
		}

		srv, err := net.ListenTCP("tcp", addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
			os.Exit(1)
		}

		for {
			conn, err := srv.Accept()
			if err != nil {
				continue
			}

			dt := time.Now().String()
			_, _ = conn.Write([]byte(dt))
			_ = conn.Close()
		}
	}

	h.Logger.Debugf("ntp listening at %d", port)

	return nil
}

func (h *Http) Dispose() error {
	return nil
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

func (h *Http) remoteTimeRouter(r chi.Router) {
	handler := func(class, method string) http.HandlerFunc {
		return h.Invoker.Execute(class, method)
	}

	r.Route(ApiObjsTime, func(r chi.Router) {
		r.Post("/", handler("TimeService", "Create"))
		r.Get("/", handler("TimeService", "List"))
		r.Route("/{id}", func(r chi.Router) {
			r.Delete("/", handler("TimeService", "Delete"))
			r.Get("/", handler("TimeService", "Get"))
			r.Get("/now", handler("TimeService", "Now"))
		})
	})
}
