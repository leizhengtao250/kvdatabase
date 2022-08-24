package skiplist

import (
	"log"
	"sync/atomic"
	"unsafe"
)

// unsafe.Sizeof(x) 来确定一个变量占用的内存字节数
const (
	MaxNodeSize = int(unsafe.Sizeof(Node{}))
	offsetSize  = int(unsafe.Sizeof(uint32(0)))
	nodeAlign   = int(unsafe.Sizeof(uint64(0))) - 1
)

type Arena struct {
	n          uint32 //已经存了n个字节
	buf        []byte //当前已经开辟的内存空间
	shouldGrow bool   //是否需要增长空间
}

//go:linkname FastRand runtime.fastrand
func FastRand() uint32

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
	l := uint32(v.EncodedSize())
	offset := a.allocate(l)        //把l大小的值放入内存中返回的offset,也就是v的起始位置
	v.EncodedValue(a.buf[offset:]) //这个值为value的size大小
	return offset
}

func (a *Arena) GetVal(offset, size uint32) (ret ValueStruct) {
	ret.DecodeValue(a.buf[offset : offset+size])
	return

}

func (a *Arena) PutKey(key []byte) uint32 {
	keysz := uint32(len(key))
	offset := a.allocate(keysz)
	buf := a.buf[offset : offset+keysz]
	AssertTrue(len(key) == copy(buf, key))
	return offset
}

/**
	就是开辟一个存放整个node节点的内存空间，使用内存对齐节省空间
**/
func (a *Arena) PutNode(height int) uint32 {
	unusedSize := (maxHeight - height) * offsetSize   //从当前层到最高层，每层都需要一个offset的大小的空间装入指针
	l := uint32(MaxNodeSize - unusedSize + nodeAlign) //做内存对齐的
	n := a.allocate(l)
	m := (n + uint32(nodeAlign)) &^ uint32(nodeAlign)
	return m
}

/**
	给内存重新分配空间
	设内存空间大小为 m ，已经分配了n ，node节点大小为k，需要加入的大小为l
	1.n+l>m-k
**/

func (a *Arena) allocate(l uint32) uint32 {
	offset := atomic.AddUint32(&a.n, l) //n+l
	if !a.shouldGrow {
		AssertTrue(int(offset) <= len(a.buf))
		return offset - l //不能分配，回到初始位置n+l-l=n
	}

	if int(offset) > len(a.buf)-MaxNodeSize { //n+l>m-k,此时需要分配
		growBy := uint32(len(a.buf))
		if growBy > 1<<30 {
			growBy = 1 << 30
		}
		if growBy < l {
			growBy = l
		}
		newBuf := make([]byte, len(a.buf)+int(growBy))
		AssertTrue(len(a.buf) == copy(newBuf, a.buf))
		a.buf = newBuf
	}
	return offset - l
}

func AssertTrue(b bool) {
	if !b {
		log.Fatalf("#{errors.Error}")
	}
}

func AssertTruef(b bool, format string, args ...interface{}) {
	if !b {
		log.Fatalf("#{errors.Errorf(format,args...)}")
	}
}

func (a *Arena) getNodeOffset(n *Node) uint32 {
	if n == nil {
		return 0
	}
	return uint32(uintptr(unsafe.Pointer(n)) - uintptr(unsafe.Pointer(&a.buf[0])))
}
