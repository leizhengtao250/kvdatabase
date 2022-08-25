package skiplist

/**
	用一个4bit 去作为缓存访问次数计数，最高15次
	1byte=uint8
**/

type cmRow []byte

/**
	1uint8 = 2 counter ---1 counter = 4 bit
	32 uint8 = 64 counter
**/
func newCmRow(numCounters int64) cmRow {
	return make(cmRow, numCounters/2)
}

/**
TODO
	0x0f=00001111
	byte(r[n/2]>>((n&1)*4)) & 0x0f


*/
func (r cmRow) get(n uint64) byte {
	return byte(r[n/2]>>((n&1)*4)) & 0x0f
}

/**
一个counter代表一个计数器
*/

func (r cmRow) increment(n uint64) {
	i := n / 2

}
