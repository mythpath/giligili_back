package util

import (
	"testing"
	"os"
	"io"
	"fmt"
	"regexp"
)

func TestReadAllDir(t *testing.T) {
	dirs, err := FindAllDir("/var/log", 1)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(dirs)
}

func TestFindAllFile(t *testing.T) {
	fns, err := FindAllFile("./", 1)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(fns)
}

func TestFindAllPaths(t *testing.T) {
	filepaths, err := FindAllPaths("./file.go", nil)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(filepaths)
}

func TestFindAllPaths_1(t *testing.T) {
	filepaths, err := FindAllPaths("./", func(path string) (bool, error) {
		match, err := regexp.MatchString("file", path)
		if err != nil {
			return false, err
		}
		return match, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(filepaths)
}

func TestDevINode(t *testing.T) {
	fi, err := os.Lstat("./file.go")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(DevINode(fi))
}

func TestReverseReadFile_ReadLine(t *testing.T) {
	f, err := OpenReverseReadFile("./t")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	for {
		line, err := f.ReadLine()
		if err == io.EOF {
			return
		}

		fmt.Println(string(line))
	}
}
