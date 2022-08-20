package skiplist

import (
	"unsafe"
)

type Arena struct {
	n   uint32 //已经存了n个字节
	buf []byte //当前已经开辟的内存空间
}

func (a *Arena) GetKey(keyOffset uint32, keySize uint16) []byte {
	return a.buf[keyOffset : keyOffset+uint32(keySize)]
}

func (a *Arena) GetNode(offset uint32) *Node {
	if offset == 0 {
		return nil
	}
	return (*Node)(unsafe.Pointer(&a.buf[offset]))
}

func (a *Arena) PutVal(v ValueStruct) uint32 {
	l := len([]byte(v))
	a.buf[n : n+l] = []byte(v)
	return a.n

}
