package store

import (
	"kvdatabase/dbfile"
	"kvdatabase/utils"
	"os"
	"sync/atomic"
)

type table struct {
	ss  *SSTable
	lm  *levelManager
	fid uint64
	ref int32 //一旦有个进程或者协程正在使用table +1
}

func OpenTable(lm *levelManager, tableName string, builder *tableBuilder) *table {
	sstSize := int(lm.opt.SSTableMaxSz)
	if builder != nil {
		sstSize = int(builder.done().size)
	}
	var (
		t   *table
		err error
	)
	fid := utils.FID(tableName)
	//对builder存在的情况，把buf flush到磁盘
	if builder != nil {
		if t, err = builder.flush(lm, tableName); err != nil {
			utils.Err(err)
			return nil
		}
	} else {
		t = &table{
			lm:  lm,
			fid: fid,
		}
		t.ss = OpenSStable(&dbfile.OptionsSST{
			FileName: tableName,
			Dir:      lm.opt.WorkDir,
			Flag:     os.O_CREATE | os.O_RDWR,
			MaxSz:    sstSize,
		})
		//正在使用sstbale
		t.IncrRef()
		//初始化sst文件，加载sst文件的索引
		if err := t.ss.Init(); err != nil {
			utils.Err(err)
			return nil
		}
		//获取sst的最大key，使用迭代器
		itr := t.NewIterator(&utils.OptionsIter{}) //默认是降序
		defer itr.Close()
		//定位到初始位置就是最大的key
		itr.Rewind()
	}

}

/**
返回的是多少程序正在并发使用table
*/

func (t *table) IncrRef() {
	atomic.AddInt32(&t.ref, 1)
}

type blockIterator struct {
	data         []byte
	idx          int
	err          error
	baseKey      []byte
	key          []byte
	val          []byte
	entryOffsets []uint32
	block        *block
	tableID      uint64
	blockID      int
	preOverlap   uint16
	it           utils.Item
}

type tableIterator struct {
	item     utils.Item
	opt      *utils.OptionsIter
	t        *table
	blockPos int
	bi       *blockIterator
	err      error
}

func (it *tableIterator) Next() {
	it.err = nil
	if it.blockPos
}

func (t *tableIterator) Vaild() bool {
	//TODO implement me
	panic("implement me")
}

func (t *tableIterator) Rewind() {
	//TODO implement me
	panic("implement me")
}

func (t *tableIterator) Item() utils.Item {
	//TODO implement me
	panic("implement me")
}

func (t *tableIterator) Close() error {
	//TODO implement me
	panic("implement me")
}

func (t *tableIterator) Seek(key []byte) {
	//TODO implement me
	panic("implement me")
}

func (t *table) NewIterator(options *utils.OptionsIter) utils.Iterator {
	//在表上迭代
	t.IncrRef()
	return &tableIterator{
		opt: options,
		t:   t,
		bi:  &blockIterator{},
	}
}
