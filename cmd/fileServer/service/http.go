package service

import (
	"github.com/go-chi/chi"
	"log"
	"net/http"
	"regexp"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging"
	"selfText/giligili_back/libcommon/net/http/invoker"
	"selfText/giligili_back/service/file/common"
	sFile "selfText/giligili_back/service/file/file"
	"selfText/giligili_back/service/file/snapshot"
	"strings"
)

type Http struct {
	Config  brick.Config           `inject:"config"`
	Logger  *logging.LoggerService `inject:"LoggerService"`
	Invoker *invoker.Invoker       `inject:"Invoker"`

	FileServer     *sFile.FileRouting       `inject:"FileServer"`
	UploadServer   *sFile.UploadServer      `inject:"UploadServer"`
	SnapshotServer *snapshot.SnapshotServer `inject:"SnapshotServer"`

	storagePath string
}

func (h *Http) Init() error {
	port := h.Config.GetMapString("http", "port")
	h.storagePath = h.Config.GetString("file_storage_path")
	if port == "" {
		h.Logger.Infof("http_port isn't setted")
	}

	resourceLock := common.NewResourceLockBucket()
	r := chi.NewRouter()
	for url, file := range h.Config.GetMap("files") {
		dir := h.replace(file.(string))

		h.Logger.Infof("file router: %s -> %s", url, dir)
		h.FileServer.Handle(r, url, sFile.Dir(dir), resourceLock)
	}

	for url, file := range h.Config.GetMap("uploads") {
		dir := h.replace(file.(string))

		h.Logger.Infof("upload router: %s -> %s", url, dir)
		h.UploadServer.Handle(r, url, dir, resourceLock)
	}

	h.SnapshotServer.Init(r, h.Config, h.replace, resourceLock)

	go func() {
		log.Fatalln(http.ListenAndServe(":"+port, r))
	}()
	return nil
}

func (h *Http) replace(template string) string {
	reg := regexp.MustCompile(`^\$\{(.+)\}`)
	match := reg.FindStringSubmatch(template)

	if len(match) < 2 {
		return template
	}

	val := h.storagePath
	if val != "" {
		return strings.Replace(template, match[0], val, 1)
	}

	return template
}
