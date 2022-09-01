package store

import (
	"kvdatabase/dbfile"
	"kvdatabase/utils"
	"os"
	"sync"
	"time"
)

//sstable

type SSTable struct {
	lock       *sync.RWMutex
	f          *dbfile.MmapFile
	maxKey     []byte
	minKey     []byte
	indexStart int
	indexLen   int
	fid        uint64
	createAt   time.Time
}

/**
打开一个sstable，具备mmap，
*/

func OpenSStable(opt *dbfile.Options) *SSTable {
	omf, err := dbfile.OpenMmapFile(opt.FileName, os.O_CREATE|os.O_RDWR, opt.MaxSz)
	utils.Err(err)
	return &SSTable{
		f:    omf,
		fid:  opt.FID,
		lock: &sync.RWMutex{},
	}
}
