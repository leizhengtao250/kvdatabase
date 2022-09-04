package test

import (
	"fmt"
	"testing"
)

func TestBaseKey(t *testing.T) {
	a := 1
	c := a
	fmt.Printf("%p\n", c)

}

var b []byte

func allocate(need int) []byte {
	a := make([]byte, need)
	b = a
	fmt.Printf("%T,%T", a, b)
	return b
}
