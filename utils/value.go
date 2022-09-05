package utils

import (
	"encoding/binary"
	"reflect"
	"unsafe"
)

func U32SliceToBytes(u32s []uint32) []byte {
	if len(u32s) == 0 {
		return nil
	}
	var b []byte
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	hdr.Len = len(u32s) * 4
	hdr.Cap = hdr.Len
	hdr.Data = uintptr(unsafe.Pointer(&u32s[0]))
	return b
}

/**
将v转换为 4个字节
*/
func U32TOBytes(v uint32) []byte {
	var uBuf [4]byte
	binary.BigEndian.PutUint32(uBuf[:], v)
	return uBuf[:]
}

func U64TOBytes(v uint64) []byte {
	var uBuf [8]byte
	binary.BigEndian.PutUint64(uBuf[:], v)
	return uBuf[:]
}

func ByteToU32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

func ByteToU64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}
