package skiplist

import (
	"fmt"
	"testing"
)

func TestA(t *testing.T) {
	//1100 1011
	a := uint32(2)
	v := a>>17 | a<<15
	fmt.Println(v)

}
