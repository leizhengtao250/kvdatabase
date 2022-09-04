package utils

import "encoding/binary"

/**

entry是在block中的一条条数据
block组成sstable
*/

type Entry struct {
	Key         []byte
	Value       []byte
	ExpiresAt   uint64
	Meta        byte
	Version     uint64
	Offset      uint32
	HeaderLen   int   //header的长度
	ValThresold int64 //阈值
}

type ValueStruct struct {
	Meta      byte
	Value     []byte
	ExpiresAt uint64
	Version   uint64 //这个版本号不是连续的，只是内部作用
}

/**
valueStruct 在编码之后需要 这么大空间存储
*/

func (vs *ValueStruct) EncodeSize() uint32 {
	sz := len(vs.Value) + 1 //meta占用1B
	enc := sizeVarint(vs.ExpiresAt)
	return uint32(sz + enc)
}

/**
将value 存入b中并且返回存入的字节数
*/

func (vs *ValueStruct) EncodeValue(b []byte) uint32 {
	b[0] = vs.Meta
	sz := binary.PutUvarint(b[1:], vs.ExpiresAt)
	n := copy(b[1+sz:], vs.Value)
	return uint32(1 + sz + n)
}

/**
Varint编码中真是数据的字节计算
*/
func sizeVarint(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
