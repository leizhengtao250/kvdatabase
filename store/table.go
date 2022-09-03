package store

import "kvdatabase/proto"

type table struct {
	ss  *SSTable
	lm  *levelManager
	fid uint64
	ref int32
}

func OpenTable(lm *levelManager, tableName string, builder *tableBuilder) *table {
	sstSize := int(lm.opt.SSTableMaxSz)
	if builder != nil {
		sstSize = int(builder)
	}
}
