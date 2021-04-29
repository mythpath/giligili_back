package snapshot

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

type File struct {
	Url     string `json:"url"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

func (f *File) String() string {
	return fmt.Sprintf("url:%s,name:%s,type:%s,content:%s", f.Url, f.Name, f.Type, f.Content)
}

func (f *File) Clone() *File {
	return &File{
		Url:     f.Url,
		Name:    f.Name,
		Type:    f.Type,
		Content: f.Content,
	}
}

func EncodeDirs(dirs []os.FileInfo, path string) ([]*File, error) {
	files := make([]*File, 0)
	for _, d := range dirs {
		name := d.Name()
		if name == "" || name == "." || name == ".." {
			continue
		}
		if name[0] == '.' {
			continue
		}
		ftype := "file"
		if d.IsDir() {
			ftype = "dir"
		}
		url := url.URL{Path: filepath.Join(path, name)}
		file := &File{Url: url.String(), Name: name, Type: ftype}
		files = append(files, file)
	}
	return files, nil
}
