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
	lock *sync.RWMutex
	f    *dbfile.MmapFile
	//为了加速查找数据，每个sst文件都会有最大key和最小key
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

//Init初始化
func (ss *SSTable)Init()error{
	var ko *
}
