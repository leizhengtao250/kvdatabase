package skiplist

import (
	"bytes"
	"encoding/binary"
	"math"
)

func SameKey(src, dst []byte) bool {
	if len(src) != len(dst) {
		return false
	}
	return bytes.Equal(ParseKey(src), ParseKey(dst))
}

/**
	为了减少比较次数
**/
func ParseKey(key []byte) []byte {
	if len(key) < 8 {
		return key
	}
	return key[:len(key)-8]
}

/**
	从key中解析出时间戳
**/

func ParseTs(key []byte) uint64 {
	if len(key) <= 8 {
		return 0
	}
	return math.MaxUint64 - binary.BigEndian.Uint64(key[len(key)-8:])
}
