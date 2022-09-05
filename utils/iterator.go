package utils

//迭代器
type Iterator interface {
	Next()
	Vaild() bool
	Rewind()
	Item() Item
	Close() error
	Seek(key []byte)
}

type Item interface {
	Entry() *Entry
}

type OptionsIter struct {
	Prefix []byte
	IsAsc  bool
}
