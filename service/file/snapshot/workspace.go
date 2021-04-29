package snapshot

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/service/file/file"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	deployFile = ".deploy"
	md5File    = ".md5"
)

type WorkspaceFactory struct {
	config brick.Config

	replaceF func(string) string
}

func NewWorkspaceFactory(config brick.Config, templateF func(string) string) *WorkspaceFactory {
	return &WorkspaceFactory{
		config:   config,
		replaceF: templateF,
	}
}

func (p *WorkspaceFactory) NewWorkspaceHandle(path string) (*WorkspaceHandle, error) {
	for url, route := range p.config.GetMap("files") {
		dir := p.replaceF(route.(string))

		if strings.HasPrefix(path, url) {
			return NewWorkspaceHandle(url, file.Dir(dir)), nil
		}
	}

	return nil, fmt.Errorf("error %s", path)
}

type WorkspaceHandle struct {
	root file.FileSystem

	prefix string
}

func NewWorkspaceHandle(prefix string, root file.FileSystem) *WorkspaceHandle {
	return &WorkspaceHandle{
		root:   root,
		prefix: prefix,
	}
}

// url is file
func (p *WorkspaceHandle) Md5(url *Url) (string, error) {
	filename := url.SubPath(string(p.prefix))
	fi, err := p.root.Stat(filename)
	if err != nil {
		logrus.WithField("filename", filename).Errorf("failed to stat file: %v", err)
		return "", err
	}
	//_, iNode := file.DevINode(fi)
	md5Filename := filepath.Join(filepath.Dir(filename), p.md5Filename(fmt.Sprintf("%s", fi.Name())))

	f, err := p.root.Open(md5Filename)
	if err != nil {
		logrus.WithField("filename", md5Filename).Errorf("failed to open md5 file: %v", err)
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	buf.ReadFrom(f)

	return strings.TrimSpace(buf.String()), nil
}

// url is file
func (p *WorkspaceHandle) WriteMD5(url *Url, md5 string) error {
	filename := url.SubPath(string(p.prefix))
	fi, err := p.root.Stat(filename)
	if err != nil {
		logrus.WithField("filename", filename).Errorf("failed to stat file: %v", err)
		return err
	}
	//_, iNode := file.DevINode(fi)

	md5Filename := filepath.Join(filepath.Dir(filename), p.md5Filename(fmt.Sprintf("%s", fi.Name())))
	md5Dirname := filepath.Dir(md5Filename)
	if err := p.root.MkdirAll(md5Dirname, os.ModePerm); err != nil {
		logrus.WithField("dirname", md5Dirname).Errorf("failed to create directory: %v", err)
		return err
	}

	f, err := p.root.OpenFile(md5Filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		logrus.WithField("filename", md5Filename).Errorf("failed to open md5 file: %v", err)
		return err
	}
	defer f.Close()

	_, err = f.WriteString(strings.TrimSpace(md5))

	return err
}

func (p *WorkspaceHandle) ReadLastDeploy(url *Url) (deployId uint, err error) {
	dir := url.SubDir(string(p.prefix))
	filename := filepath.Join(dir, p.deployFilename(url))
	f, err := p.root.Open(filename)
	if err != nil {
		logrus.WithField("filename", filename).Errorf("failed to open file: %v", err)
		return 0, err
	}
	defer f.Close()

	var buf bytes.Buffer
	buf.ReadFrom(f)

	var id int
	id, err = strconv.Atoi(buf.String())
	deployId = uint(id)
	return
}

func (p *WorkspaceHandle) Open(url *Url) (http.File, error) {
	subPath := url.SubPath(p.prefix)

	return p.root.Open(subPath)
}

func (p *WorkspaceHandle) Close(f http.File) {
	f.Close()
}

func (p *WorkspaceHandle) WriteDeploy(url *Url, deployId uint) error {
	filename := filepath.Join(url.SubDir(p.prefix), p.deployFilename(url))
	dir := filepath.Dir(filename)
	if err := p.root.MkdirAll(dir, os.ModePerm); err != nil {
		logrus.WithField("dir", dir).Errorf("failed to make dir: %v", err)
		return err
	}
	f, err := p.root.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		logrus.WithField("filename", filename).Errorf("failed to open file: %v", err)
		return err
	}
	defer f.Close()

	reader := strings.NewReader(fmt.Sprintf("%d", deployId))
	_, err = io.Copy(f, reader)

	return err
}

func (p *WorkspaceHandle) CopyFileFromSnapshots(url *Url, reader io.Reader) error {
	filename := filepath.Join(url.SubPath(p.prefix))
	dirname := filepath.Dir(filename)
	if err := p.root.MkdirAll(dirname, os.ModePerm); err != nil {
		logrus.WithField("dirname", dirname).Errorf("failed to create directory: %v", err)
		return err
	}

	f, err := p.root.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		logrus.WithField("filename", filename).Errorf("failed to open file: %v", err)
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, reader)

	return err
}

func (p *WorkspaceHandle) md5Filename(iNode string) string {
	return filepath.Join(md5File, iNode)
}

func (p *WorkspaceHandle) deployFilename(url *Url) string {
	return filepath.Join(deployFile, fmt.Sprintf("%s%s", url.Base(), deployFile))
}
