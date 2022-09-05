package test

import (
	"fmt"
	"kvdatabase/dbfile"
	"os"
	"testing"
)

func TestBaseKey(t *testing.T) {
	a := "/Users/hello/project/golanguage/sst/test/a.txt"
	f, _ := os.Open(a)
	m := &dbfile.MmapFile{
		Data: make([]byte, 4),
		Fd:   f,
	}
	b := make([]byte, 4)
	b = []byte("abcd")
	dst, _ := m.Bytes(0, 4)
	copy(dst, b)

}

var b []byte

func allocate(need int) []byte {
	a := make([]byte, need)
	b = a
	fmt.Printf("%T,%T", a, b)
	return b
}
