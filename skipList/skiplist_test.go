package skiplist

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestA(t *testing.T) {
	fmt.Println(unsafe.Sizeof(Example1{}))
	fmt.Println(unsafe.Sizeof(Example2{}))
	fmt.Println(unsafe.Alignof(Example1{}))
	fmt.Println(unsafe.Alignof(Example2{}))
}

type Example1 struct {
	a int32
	b int64
	c int32
}

type Example2 struct {
	a int32
	b int32
	c int64
}
