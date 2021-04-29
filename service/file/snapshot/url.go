package snapshot

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// Url /3/nerv-app/2/api/scripts/nerv-app/type.json
type Url struct {
	deployId uint
	name     string
	path     string

	sep uint
	url string
}

func NewUrl(path string) (*Url, error) {
	url := &Url{
		url: path,
	}

	dirs := strings.Split(path, "/")
	if len(dirs) < 2 {
		return nil, fmt.Errorf("invalid path: %s", path)
	}

	var sepCount int
	sepCount, err := strconv.Atoi(dirs[1])
	if err != nil {
		return url, err
	}

	if sepCount <= 0 || sepCount >= len(dirs) {
		return url, fmt.Errorf("error %s", path)
	}
	url.sep = uint(sepCount)

	for i := 2; i < sepCount; i++ {
		url.name = filepath.Join(url.name, dirs[i])
	}

	id, err := strconv.Atoi(dirs[sepCount])
	if err != nil {
		return url, fmt.Errorf("error parse %s", path)
	}
	url.deployId = uint(id)

	upath := "/"
	for i := sepCount + 1; i < len(dirs); i++ {
		upath = filepath.Join(upath, dirs[i])
	}
	url.path = upath

	return url, nil
}

func (u *Url) UrlPath() string {
	return u.url
}

func (u *Url) MetaPath() string {
	return filepath.Join("/", u.name, fmt.Sprintf("%d", u.deployId))
}

func (u *Url) Snapshot() string {
	return filepath.Join(fmt.Sprintf("/%d", u.sep), u.name, fmt.Sprintf("%d", u.deployId))
}

// RealPath ---> /api/scripts/nerv-app/type.json
func (u *Url) RealPath() string {
	return u.path
}

// SubPath ---> /nerv-app/type.json
func (u *Url) SubPath(prefix string) string {
	index := strings.Index(u.path, prefix)
	if index == -1 {
		return u.path
	}

	return strings.TrimPrefix(string(u.path)[index:], prefix)
}

// SubDir ---> /nerv-app
func (u *Url) SubDir(prefix string) string {
	subPath := u.SubPath(prefix)
	if subPath == "" {
		return filepath.Dir(u.path)
	}

	return filepath.Dir(subPath)
}

// Path ---> /nerv-app/2/api/scripts/nerv-app/type.json
func (u *Url) Path() string {
	return filepath.Join("/", u.name, fmt.Sprintf("%d", u.deployId), u.path)
}

// Base ---> /nerv-app/2/api/scripts/nerv-app
func (u *Url) Base() string {
	return filepath.Base(u.Path())
}

// ReplaceUrl ---> /nerv-app/$deployId/api/scripts/nerv-app/type.json
func (u *Url) ReplaceUrl(deployId uint) (*Url, error) {
	url := u.Clone()

	url.deployId = deployId
	url.url = filepath.Join(fmt.Sprintf("/%d", url.sep), url.name, fmt.Sprintf("%d", url.deployId), url.path)

	return url, nil
}

func (u *Url) String() string {
	return fmt.Sprintf("<sep:%d,name:%s,deployId:%d,path:%s,url:%s>", u.sep, u.name, u.deployId, u.path, u.url)
}

func (u *Url) Clone() *Url {
	return &Url{
		url:      u.url,
		path:     u.path,
		name:     u.name,
		sep:      u.sep,
		deployId: u.deployId,
	}
}
