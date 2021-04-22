package file

import (
	"io"
	"path/filepath"
	"fmt"
	"os"
	"crypto/md5"
	"bytes"
)

const (
	md5File = ".md5"
)

// reader is file content
func WriteMD5(filename string, reader io.Reader) error {
	dir := filepath.Join(filepath.Dir(filename), md5File)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}

	h := md5.New()
	_, err = io.Copy(h, reader)
	if err != nil {
		return err
	}

	filename = filepath.Join(dir, fmt.Sprintf("%s%s", filepath.Base(filename), md5File))
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, bytes.NewReader([]byte(fmt.Sprintf("%x", h.Sum(nil)))))

	return err
}

// reader is hash.Sum(nil) value
func WriteMD5Value(filename string, reader io.Reader) error {
	dir := filepath.Join(filepath.Dir(filename), md5File)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}

	iNode, err := INode(filename)
	if err != nil {
		return err
	}

	filename = filepath.Join(dir, fmt.Sprintf("%d", iNode))
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, reader)

	return err
}
