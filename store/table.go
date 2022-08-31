package store

type table struct {
	ss  *sstable
	lm  *levelManager
	fid uint64
	ref int32
}

func OpenTable(lm *levelManager, tableName string, builder *tableBuilder) *table {

}
