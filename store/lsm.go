package store

type OptionLsm struct {
	WorkDir            string  //文件存储的目录
	MemTableSize       int64   //跳表的最大占用空间
	SSTableMaxSz       int64   //sst的最大占用空间
	BlockSize          int     //每个block的大小
	BloomFalsePositive float64 //布隆过滤器的假阳性大小
}
