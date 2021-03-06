package util

import (
	"os"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

func mmap(f *os.File, length int) ([]byte, error) {
	return unix.Mmap(int(f.Fd()), 0, length, unix.PROT_READ, unix.MAP_SHARED)
}

func munmap(b []byte) (err error) {
	return unix.Munmap(b)
}

type MmapFile struct {
	f *os.File
	b []byte
}

func OpenMmapFile(path string) (*MmapFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "try lock file")
	}
	info, err := f.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "stat")
	}

	b, err := mmap(f, int(info.Size()))
	if err != nil {
		return nil, errors.Wrap(err, "mmap")
	}

	return &MmapFile{f: f, b: b}, nil
}

func (f *MmapFile) Close() error {
	err0 := munmap(f.b)
	err1 := f.f.Close()

	if err0 != nil {
		return err0
	}
	return err1
}

func (f *MmapFile) File() *os.File {
	return f.f
}

func (f *MmapFile) Bytes() []byte {
	return f.b
}
