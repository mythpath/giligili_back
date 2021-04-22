package util

import (
	"bufio"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func EnsureDir(dir string) error {
	if IsExist(dir) {
		return nil
	}

	return os.MkdirAll(dir, os.ModePerm)
}

func IsExist(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}

func ReadLine(r *bufio.Reader) ([]byte, error) {
	line, isPrefix, err := r.ReadLine()
	for isPrefix && err == nil {
		var bs []byte
		bs, isPrefix, err = r.ReadLine()
		line = append(line, bs...)
	}

	return line, err
}

func FindAllDirNames(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	dirs := make([]string, 0)
	if fi.IsDir() {
		fis, err := f.Readdir(-1)
		if err != nil {
			return nil, err
		}

		// read dir
		for _, fi := range fis {
			if fi.IsDir() {
				if fi.Name() == "." || fi.Name() == ".." {
					continue
				}
				dirs = append(dirs, fi.Name())
			}
		}

		return dirs, nil
	}

	dirs = append(dirs, filepath.Dir(path))
	return dirs, nil
}

// FindAllDir is find all dir names within path
func FindAllDir(path string, depth int) ([]string, error) {

	return findAllDir(path, 0, depth)
}

// findAllDir finds all dir name in assign path
func findAllDir(path string, depth int, total int) ([]string, error) {
	// check depth
	if total != -1 && depth >= total {
		return nil, nil
	}
	depth += 1

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	dirs := make([]string, 0)
	if fi.IsDir() {
		fis, err := f.Readdir(-1)
		if err != nil {
			return nil, err
		}

		// read dir
		for _, fi := range fis {
			if fi.IsDir() {
				dirs = append(dirs, filepath.Join(path, fi.Name()))

				dr, err := findAllDir(filepath.Join(path, fi.Name()), depth, total)
				if err != nil {
					return nil, err
				}

				dirs = append(dirs, dr...)
			}
		}

		return dirs, nil
	}

	dirs = append(dirs, filepath.Dir(path))
	return dirs, nil
}

func FindAllDirPaths(dir string, match func(path string) (bool, error)) ([]string, error) {
	filepaths := make([]string, 0, 1)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		if match != nil {
			ok, err := match(path)
			if err != nil {
				return err
			}

			if ok {
				filepaths = append(filepaths, path)
			}
			return nil
		}

		filepaths = append(filepaths, path)
		return nil
	})

	return filepaths, err
}

func FindAllPaths(dir string, match func(path string) (bool, error)) ([]string, error) {
	filepaths := make([]string, 0, 1)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if match != nil {
			ok, err := match(path)
			if err != nil {
				return err
			}

			if ok {
				filepaths = append(filepaths, path)
			}
			return nil
		}

		filepaths = append(filepaths, path)
		return nil
	})

	return filepaths, err
}

// FindAllFile is find all file names within path
func FindAllFile(path string, depth int) ([]string, error) {

	return findAllFile(path, 0, depth, nil)
}

func FindAllFileFunc(path string, depth int, f func(path string) bool) ([]string, error) {
	return findAllFile(path, 0, depth, f)
}

func findAllFile(path string, depth int, total int, match func(path string) bool) ([]string, error) {
	// check depth
	if total != -1 && depth >= total {
		return nil, nil
	}
	depth += 1

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	filenames := make([]string, 0)
	if fi.IsDir() {
		fis, err := f.Readdir(-1)
		if err != nil {
			return nil, err
		}

		for _, fi := range fis {
			if fi.IsDir() {
				fns, err := findAllFile(filepath.Join(path, fi.Name()), depth, total, match)
				if err != nil {
					return nil, err
				}

				filenames = append(filenames, fns...)
				continue
			}

			if strings.HasPrefix(fi.Name(), ".") {
				continue
			}

			if match != nil {
				if match(fi.Name()) {
					filenames = append(filenames, filepath.Join(path, fi.Name()))
				}
				continue
			}

			filenames = append(filenames, filepath.Join(path, fi.Name()))
		}

		return filenames, nil
	}

	return append(filenames, path), nil
}

// realPath handle symlink
func RealPath(path string) (string, error) {
	newPath := path
	fi, err := os.Lstat(path)
	if err != nil {
		return "", err
	}

	if fi.Mode()&os.ModeSymlink != 0 {

		newPath, err = filepath.EvalSymlinks(path)
		if err != nil {
			return "", err
		}
	}

	newPath, err = filepath.Abs(newPath)
	return newPath, err
}

// DevINode returns fi.Dev+fi.Ino
func DevINode(fi os.FileInfo) (uint64, uint64) {
	s := fi.Sys()
	if s == nil {
		return 0, 0
	}

	switch s := s.(type) {
	case *syscall.Stat_t:
		return uint64(s.Dev), s.Ino
	}
	return 0, 0
}

// DirName returns dir and file name
func DirFilename(path string) (dir string, fn string, err error) {
	var fi os.FileInfo

	fi, err = os.Lstat(path)
	if err != nil {
		return
	}

	dir = filepath.Dir(path)
	fn = filepath.Base(path)

	if fi.IsDir() {
		dir = path
		fn = ""
	}

	return
}

// find file by dev+iNode
func FindFile(dir string, dev uint64, iNode uint64) (os.FileInfo, error) {
	filenames, err := FindAllFile(dir, 1)
	if err != nil {
		return nil, err
	}

	for _, filename := range filenames {
		fi, err := os.Lstat(filename)
		if err != nil {
			return nil, err
		}

		dv, id := DevINode(fi)
		if dv == dev && id == iNode {
			return fi, nil
		}
	}

	return nil, nil
}

// getLatestFile
func getLatestFile(files []string) (os.FileInfo, error) {
	return getMaxFile(files, noCondition, modTimeLater)
}

// getOldestFile
func getOldestFile(files []string) (os.FileInfo, error) {
	return getMinFile(files, noCondition, modTimeLater)
}

// getMaxFile in assign condition
// gte f1 >= f2 return true
func getMaxFile(files []string, condition func(os.FileInfo) bool, gte func(f1, f2 os.FileInfo) bool) (chosen os.FileInfo, err error) {
	for _, f := range files {
		fi, err := os.Lstat(f)
		if err != nil {
			return nil, err
		}

		if condition == nil || !condition(fi) {
			continue
		}
		if chosen == nil || gte(fi, chosen) {
			chosen = fi
		}
	}
	if chosen == nil {
		return nil, os.ErrNotExist
	}
	return
}

// getMinFile 于getMaxFile 相反，返回最小的文件
func getMinFile(files []string, condition func(os.FileInfo) bool, gte func(f1, f2 os.FileInfo) bool) (os.FileInfo, error) {
	return getMaxFile(files, condition, func(f1, f2 os.FileInfo) bool {
		return !gte(f1, f2)
	})
}

// noCondition 无限制条件
func noCondition(f os.FileInfo) bool {
	return true
}

func andCondition(f1, f2 func(os.FileInfo) bool) func(os.FileInfo) bool {
	return func(fi os.FileInfo) bool {
		return f1(fi) && f2(fi)
	}
}

func orCondition(f1, f2 func(os.FileInfo) bool) func(os.FileInfo) bool {
	return func(fi os.FileInfo) bool {
		return f1(fi) || f2(fi)
	}
}

func notCondition(f1 func(os.FileInfo) bool) func(os.FileInfo) bool {
	return func(fi os.FileInfo) bool {
		return !f1(fi)
	}
}

// modTimeLater
func modTimeLater(f1, f2 os.FileInfo) bool {
	if f1.ModTime().UnixNano() != f2.ModTime().UnixNano() {
		return f1.ModTime().UnixNano() > f2.ModTime().UnixNano()
	}
	return f1.Name() > f2.Name()
}

type ReverseReadFile struct {
	filepath string

	mmapFile *MmapFile

	b      []byte
	last   int64
	offset int64
	size   int64
}

func OpenReverseReadFile(filepath string) (*ReverseReadFile, error) {
	rr := &ReverseReadFile{
		filepath: filepath,
	}

	f, err := OpenMmapFile(filepath)
	if err != nil {
		return nil, errors.Wrap(err, "open mmap file")
	}

	rr.mmapFile = f
	rr.b = f.Bytes()
	rr.size = int64(len(rr.b))
	rr.offset = rr.size
	rr.last = rr.offset
	rr.offset--

	return rr, nil
}

func (r *ReverseReadFile) Close() error {
	return r.mmapFile.Close()
}

func (r *ReverseReadFile) ReadLine() ([]byte, error) {
	line, err := r.ReadSlice('\n')
	if err != nil {
		return nil, err
	}

	if len(line) == 0 {
		return nil, nil
	}

	if line[len(line)-1] == '\n' {
		drop := 1
		if len(line) > 1 && line[len(line)-2] == '\r' {
			drop = 2
		}
		line = line[:len(line)-drop]
	}

	return line, nil
}

func (r *ReverseReadFile) ReadSlice(delim byte) ([]byte, error) {
	var line []byte

	if r.offset == 0 {
		return nil, io.EOF
	}

	for r.offset > 0 {
		r.offset--

		if r.b[r.offset] == delim {
			line = r.b[r.offset+1 : r.last]

			r.last = r.offset

			break
		}
	}

	if r.offset == 0 {
		line = r.b[r.offset:r.last]

		r.last = r.offset
		return line, nil
	}

	return line, nil
}
