package skiplist

type Entry struct {
	key []byte
}

type ValueStruct struct {
	Meta      byte //TODO 记录某种状态的
	Value     []byte
	ExpiresAt uint64
}

func (v *ValueStruct) EncodedSize() uint32 {
	sz := len(v.Value) + 1
	enc := sizeVarint(v.ExpiresAt)
	return uint32(sz + enc)
}

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
