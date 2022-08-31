package store

type OptionLsm struct {
	WorkDir            string
	MemTableSize       int64
	SSTableMaxSz       int64
	BlockSize          int
	BloomFalsePositive float64
}
