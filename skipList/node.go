package skiplist

import (
	"sync/atomic"
)

type Node struct {
	value     uint64
	keyOffset uint32
	keySize   uint16
	height    int
	tower     [maxHeight]uint32
}

/**
当前节点中取出value->这个value指的是内存管理器中的offset，size
先拿出节点中value，再拆分value，去内存管理器中取出真正的value值
*/
func (n *Node) getValueOffset() (uint32, uint32) {
	value := atomic.LoadUint64(&n.value)
	return decodeValue(value)
}

func decodeValue(v uint64) (uint32, uint32) {
	valueOffset := uint32(v)
	valueSize := uint32(v >> 32)
	return valueOffset, valueSize
}

/**
get key from arena

*/
func (n *Node) key(arena Arena) []byte {
	return arena.GetKey(n.keyOffset, n.keySize)
}

/**
set v in node(not value in arena)
*/
func (n *Node) setValue(arena *Arena, v uint64) {
	atomic.StoreUint64(&n.value, v)
}

/**
get node.next on level h  And it is a pointer
*/
func (n *Node) getNextOffset(h int) uint32 {
	return atomic.LoadUint32(&n.tower[h])
}

/**
update pointer next use cas
*/
func (n *Node) casNextOffset(h int, old, new uint32) bool {
	return atomic.CompareAndSwapUint32(&n.tower[h], old, new)
}

func newNode(arena Arena, key []byte, v ValueStruct, height int) {
	nodeOffset := 
}
