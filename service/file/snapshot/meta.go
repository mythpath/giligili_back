package snapshot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"selfText/giligili_back/service/file/file"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	metaSnapshotPrefix = "meta_snapshot_"
	metaLock           = ".locked"

	metaDirname  = ".metadir"
	metaTypeFile = "file"
	metaTypeDir  = "dir"
)

type Meta interface {
	// file type is file or dir
	Type() string
	// string
	String() string
}

// dir snapshot
type DirMetaSnapshot struct {
	Url   string
	Files []*File
}

func (d *DirMetaSnapshot) Type() string {
	return metaTypeDir
}

func (d *DirMetaSnapshot) String() string {
	return fmt.Sprintf("url: %s, files: %v", d.Url, d.Files)
}

func (d *DirMetaSnapshot) Clone() *DirMetaSnapshot {
	dm := &DirMetaSnapshot{
		Url: d.Url,
	}

	for _, fl := range d.Files {
		dm.Files = append(dm.Files, fl.Clone())
	}

	return dm
}

// file snapshot
type MetaSnapshot struct {
	RefUrl      string            // reference url of target file
	MD5         string            // the md5 value for target file
	Description map[string]string // description for file
}

func (s *MetaSnapshot) Type() string {
	return metaTypeFile
}

func (s *MetaSnapshot) Clone() *MetaSnapshot {
	ms := &MetaSnapshot{
		MD5:         s.MD5,
		RefUrl:      s.RefUrl,
		Description: make(map[string]string),
	}

	for k, v := range s.Description {
		ms.Description[k] = v
	}

	return ms
}

func (s *MetaSnapshot) String() string {
	return fmt.Sprintf("refUrl:%s,md5:%s,Desc:%v",
		s.RefUrl, s.MD5, s.Description)
}

type MetaHandle struct {
	root file.FileSystem
}

func NewMetaSnapshotHandle(root file.FileSystem) *MetaHandle {
	return &MetaHandle{
		root: root,
	}
}

func (p *MetaHandle) Locked(dir string) (bool, error) {
	filename := filepath.Join(dir, metaLock)
	_, err := p.root.Stat(filename)
	if err != nil && os.IsNotExist(err) {
		return false, nil
	}

	return true, nil
}

func (p *MetaHandle) Lock(dir string) error {
	filename := filepath.Join(dir, metaLock)

	f, err := p.root.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return nil
}

// remove meta
func (p *MetaHandle) RemoveMeta(filename string, meta Meta) error {
	filename = filepath.Join(p.filenameM(filename))
	if meta.Type() == metaTypeDir {
		filename = filepath.Join(filename, metaDirname)
	}

	fi, err := p.root.Stat(filename)
	if err == nil && fi != nil {
		if err := p.root.Remove(filename); err != nil {
			logrus.WithField("filename", filename).Errorf("failed to remove meta file: %v", err)
			return err
		}
	}

	return nil
}

// rewrite meta data
func (p *MetaHandle) WriteM(filename string, meta Meta) error {
	filename = filepath.Join(p.filenameM(filename))
	if meta.Type() == metaTypeDir {
		filename = filepath.Join(filename, metaDirname)
	}

	// remove meta
	fi, err := p.root.Stat(filename)
	if err == nil && fi != nil {
		if err := p.root.Remove(filename); err != nil {
			logrus.WithField("filename", filename).Errorf("failed to remove meta file: %v", err)
			return err
		}
	}

	dirPath := filepath.Dir(filename)
	if err := p.root.MkdirAll(dirPath, os.ModePerm); err != nil {
		return err
	}

	f, err := p.root.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		logrus.WithField("filename", filename).Errorf("failed to open the meta file: %v", err)
		return err
	}
	defer f.Close()

	var b []byte
	b, err = json.MarshalIndent(meta, "", "\t")
	if err != nil {
		logrus.WithField("filename", filename).Errorf("failed to marshal the meta content: %v", err)
		return err
	}
	_, err = f.Write(b)

	return err
}

// /nerv/2/api/scripts/nerv-app/type.json
// 		--->	/nerv/2/api/scripts/nerv-app/type.json
// /nerv/2/api/scripts/nerv-app
// 		--->	/nerv/2/api/scripts/nerv-app
func (p *MetaHandle) ReadM(filename string) (Meta, error) {
	filename = p.filenameM(filename)
	f, err := p.root.Open(filename)
	if err != nil {
		return nil, err
	}
	// maybe repeat closed
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		logrus.WithField("filename", filename).Errorf("failed to stat file: %v", err)
		return nil, err
	}
	isFile := true
	if fi.IsDir() {
		f.Close()
		f, err = p.root.Open(filepath.Join(filename, metaDirname))
		if err != nil {
			return nil, err
		}
		isFile = false
	}

	var b bytes.Buffer
	_, err = b.ReadFrom(f)
	if err != nil {
		return nil, err
	}

	var e error
	if isFile {
		var ms MetaSnapshot
		if e = json.Unmarshal(b.Bytes(), &ms); e == nil {
			return &ms, nil
		}
	} else {
		var dms DirMetaSnapshot
		if e = json.Unmarshal(b.Bytes(), &dms); e == nil {
			return &dms, nil
		}
	}

	return nil, e
}

func (p *MetaHandle) CopyFromWorkspace(url *Url, wsh *WorkspaceHandle) error {
	filename := p.filenameM(url.Path())
	if err := p.root.MkdirAll(filename, os.ModePerm); err != nil {
		return err
	}

	f, err := p.root.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
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
	return err
}

func (p *MetaHandle) RefFilename(filename string, ms *MetaSnapshot) string {
	if ms.RefUrl == "" {
		return filename
	}

	return ms.RefUrl
}

func (p *MetaHandle) ReadFiles(path string) ([]string, error) {
	dir := filepath.Join(path)
	f, err := p.root.Open(dir)
	if err != nil && os.IsNotExist(err) {
		return []string{}, nil
	}

	if err != nil {
		return []string{}, nil
	}
	defer f.Close()

	filenames := make([]string, 0)
	fis, err := f.Readdir(0)
	if err != nil {
		return filenames, err
	}

	for _, fi := range fis {
		if fi.IsDir() {
			files, err := p.ReadFiles(filepath.Join(path, fi.Name()))
			if err != nil {
				return filenames, err
			}

			filenames = append(filenames, files...)
			continue
		}

		if strings.HasSuffix(fi.Name(), metaLock) {
			continue
		}

		if strings.HasPrefix(fi.Name(), ".") {
			continue
		}
		filenames = append(filenames, filepath.Join(path, fi.Name()))
	}

	return filenames, nil
}

func (p *MetaHandle) CopyContentMeta(dst string, src string) error {
	filenames, err := p.ReadFiles(src)
	if err != nil {
		return err
	}

	softLinkF := func(filename string) <-chan result {
		rs := make(chan result, 1)

		go func() {
			relative := strings.TrimPrefix(filename, src)
			dstname := filepath.Join(dst, relative)
			if err := p.root.MkdirAll(filepath.Dir(dstname), os.ModePerm); err != nil {
				rs <- result{name: filename, err: err}
				close(rs)

				return
			}
			df, err := p.root.OpenFile(dstname, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				rs <- result{name: filename, err: err}
				close(rs)

				logrus.WithField("filename", dstname).Errorf("failed to open target file: %v", err)

				return
			}

			srcname := filepath.Join(filename)
			sf, err := p.root.Open(srcname)
			if err != nil {
				rs <- result{name: filename, err: err}
				close(rs)

				logrus.WithField("filename", srcname).Errorf("failed to open source file: %v", err)

				return
			}

			if _, err = io.Copy(df, sf); err != nil {
				rs <- result{name: filename, err: err}
				close(rs)

				logrus.WithFields(logrus.Fields{
					"source": srcname,
					"target": dstname,
				}).Errorf("failed to copy content: %v", err)

				return
			}

			rs <- result{name: filename, err: nil}
			close(rs)
		}()

		return rs
	}

	results := make([]<-chan result, 0, len(filenames))
	for _, filename := range filenames {
		results = append(results, softLinkF(filename))
	}

	for _, chanR := range results {
		select {
		case rs := <-chanR:
			if rs.err != nil {
				if err == nil {
					err = rs.err
				}

				logrus.WithFields(logrus.Fields{
					"source": rs.name,
					"target": dst,
				}).Errorf("failed to copy meta content: %v", rs.err)
			}
		}
	}

	return nil
}

func (p *MetaHandle) HardLinkMeta(dst string, src string) error {
	filenames, err := p.ReadFiles(src)
	if err != nil {
		return err
	}

	softLinkF := func(filename string) <-chan result {
		rs := make(chan result, 1)

		go func() {
			relative := strings.TrimPrefix(filename, src)
			dstname := filepath.Join(dst, relative)
			if err := p.root.MkdirAll(filepath.Dir(dstname), os.ModePerm); err != nil {
				rs <- result{name: filename, err: err}
				close(rs)

				return
			}

			// hard code
			dstname = filepath.Join(string(p.root.(file.Dir)), dstname)
			srcname := filepath.Join(string(p.root.(file.Dir)), filename)
			if err := os.Link(srcname, dstname); err != nil {
				rs <- result{name: filename, err: err}
				close(rs)

				return
			}

			rs <- result{name: filename, err: nil}
			close(rs)
		}()

		return rs
	}

	results := make([]<-chan result, 0, len(filenames))
	for _, filename := range filenames {
		results = append(results, softLinkF(filename))
	}

	for _, chanR := range results {
		select {
		case rs := <-chanR:
			if rs.err != nil {
				if err == nil {
					err = rs.err
				}

				logrus.WithFields(logrus.Fields{
					"source": rs.name,
					"target": dst,
				}).Errorf("failed to hard link meta: %v", rs.err)
			}
		}
	}

	return err
}

func (p *MetaHandle) targetFilename(msFilename string) string {
	return msFilename[strings.Index(msFilename, metaSnapshotPrefix)+1:]
}

func (p *MetaHandle) filenameM(filename string) string {
	return filepath.Join(filepath.Dir(filename), fmt.Sprintf("%s", filepath.Base(filename)))
}
