package service

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
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

		h.Logger.Debugf("ntp listening at %d", port)
	} else {
		//addr, err := net.ResolveUDPAddr("tcp", ":"+string(rune(port)))
		//if err != nil {
		//	fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		//	os.Exit(1)
		//}

		srv, err := net.ListenUDP("udp", &net.UDPAddr{
			IP:   net.IPv4(127, 0, 0, 1),
			Port: port,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
			os.Exit(1)
		}
		defer srv.Close()
		h.Logger.Debugf("ntp listening at %d", port)

		for {
			data := make([]byte, 4096)
			read, remoteAddr, err := srv.ReadFromUDP(data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
				continue
			}
			h.Logger.WithFields(logrus.Fields{
				"read":       read,
				"remoteAddr": remoteAddr,
			}).Infoln("read from udp")
			h.Logger.WithField("data", data).Infoln("content.")

			dt := time.Now().String()
			sendData := []byte(dt)
			_, err = srv.WriteToUDP(sendData, remoteAddr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
				return err
			}
		}
	}

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
