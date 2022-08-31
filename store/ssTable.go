package store

import "sync"

//sstableW

type sstable struct {
	lock   *sync.RWMutex
	f      *MmapFile
	maxKey []byte
	minKey []byte
}
