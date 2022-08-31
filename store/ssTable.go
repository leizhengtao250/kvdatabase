package store

import "sync"

//sstable

type sstable struct {
	lock   *sync.RWMutex
	f      *MmapFile
	maxKey []byte
	minKey []byte
}
