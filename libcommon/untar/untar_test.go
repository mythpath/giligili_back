package untar

import (
	"testing"
	"os"
)

func TestSymlink(t *testing.T) {
	reader, err := os.Open("/Users/darren/Downloads/nerv-cloud-1.6.0-darwin-x86_64.tgz")
	if err != nil {
		t.Fatal(err)
	}

	if err := Untar(reader, "./"); err != nil {
		t.Fatal(err)
	}

	//1000 0000 0000 0000 0000 0000 0000
	//1000 0000 0000 0000 0001 1110 1101
}
