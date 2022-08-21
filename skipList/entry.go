package skiplist

import "encoding/binary"

type Entry struct {
	key []byte
}

type ValueStruct struct {
	Meta      byte   //TODO 记录某种状态的
	Value     []byte //真实的value值
	ExpiresAt uint64 //记录过期时间的，这个在lsm的结构中有重要作用
}

func (v *ValueStruct) EncodedSize() uint32 {
	sz := len(v.Value) + 1
	enc := sizeVarint(v.ExpiresAt)
	return uint32(sz + enc)
}

/**
	TODO 为什么向右偏移7个单位
	因为涉及到Varint编码 每个字节只有7位是有效位，所以就是移动7位作为一个字节
	在还原过期时间时还要做处理

**/
func sizeVarint(x uint64) (n int) {
	for {
		n++
		x = x >> 7
		if x == 0 {
			break
		}
	}
	return n
}

func (v *ValueStruct) EncodedValue(b []byte) uint32 {
	b[0] = v.Meta
	sz := binary.PutUvarint(b[1:], v.ExpiresAt)
	n := copy(b[1+sz:], v.Value)
	return uint32(1 + sz + n)
}
