package file

import (
	"os"
	"syscall"
)

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

// get file inode
func INode(filename string) (uint64, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}

	_, iNode := DevINode(fi)
	return iNode, nil
}
