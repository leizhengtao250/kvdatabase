package skiplist

import "math"

/**
	指定 false positive ->fp
	其实就是允许出错的概率，给定一个不存在的值，但是经过hash运算仍在数组内的概率
	返回的是 m/n
**/
func BitsPerkey(numEntries int, fp float64) int {
	//首先根据公式计算位数组的位数
	size := -1 * float64(numEntries) * math.Log(fp) / math.Pow(0.69314718056, 2)
	//向上取整，计算的是 m/n:m是位数组的大小，n是元素的个数
	locs := math.Ceil(size / float64(numEntries))
	return int(locs)
}

func newFilter(numEntries int, fp float64) *BloomFilter {
	bitsPerkey := BitsPerkey(numEntries, fp)
	return initFilter(numEntries, bitsPerkey)
}

func initFilter(numEntries int, bitPerkey int) *BloomFilter {
	bf := &BloomFilter{}
	if bitPerkey < 0 {
		bitPerkey = 0
	}
	k := uint32(0.69 * float64(bitPerkey))
	if k < 1 {
		k = 1
	}
	if k > 30 {
		k = 30
	}
	bf.k = uint8(k)
	nBits := numEntries * int(bitPerkey)
	if nBits < 64 {
		nBits = 64
	}
	nBytes := (nBits + 7) / 8
	nBits = nBytes * 8
	filter := make([]byte, nBytes+1)

	filter[nBytes] = uint8(k)
	bf.bitmap = filter
	return bf
}

/**
根据公式计算hash函数的数量=0.7*(m/n)
*/

func CalcHashNum(bitPerkey int) (k uint32) {
	k = uint32(0.69 * float64(bitPerkey))
	if k < 1 {
		k = 1
	}
	if k > 30 {
		k = 30
	}
	return
}

/**
	这个appendFilter1里面每个keys在过滤器中占8个字节
	实际上使用位图法只需要1个bit位就可以了
**/

func appendFilter1(keys []uint32, bitPerkey int) []byte {
	if bitPerkey < 0 {
		bitPerkey = 0
	}
	//hash函数的数量
	k := CalcHashNum(bitPerkey)
	//将keys中key放入到布隆过滤器中
	// 位数组大小m，元素个数为n
	/**
		bitPerkey = m/n
		len(keys)=n
		len(keys)*(bitPerkey)=m ----nBits位数组大小

	**/
	nBits := len(keys) * int(bitPerkey)
	//位图数组
	filter := make([]byte, nBits)
	//key->hash->位图数组
	for _, h := range keys {
		delta := h>>17 | h<<15 //把key转化成32位
		for j := uint32(0); j < k; j++ {
			bitPos := h % uint32(nBits) //这个地方就属于hash函数
			filter[bitPos] = 1
			h += delta
		}
	}
	return filter
}

func appendFilter(buf []byte, keys []uint32, bitsPerkey int) []byte {
	if bitsPerkey < 0 {
		bitsPerkey = 0
	}
	k := CalcHashNum(bitsPerkey)
	nBits := len(keys) * int(bitsPerkey)
	if nBits < 64 {
		nBits = 64
	}

	nBytes := (nBits + 7) / 8 //把位图按照每8位一组，共nBytes组
	nBits = nBytes * 8        //每组8位，共nBits位

	filter := make([]byte, nBytes+1) //分为这么多组
	for _, h := range keys {
		delta := h>>17 | h<<15
		for j := uint32(0); j < k; j++ {
			bitPos := h % uint32(nBits)
			filter[bitPos/8] |= 1 << (bitPos % 8)
			h += delta
		}
	}
	return filter
}

type Filter []byte
type BloomFilter struct {
	bitmap Filter
	k      uint8 //hash 函数个数
}

//将经过hash计算之后的值插入到位图中

func (f *BloomFilter) Insert(h uint32) bool {
	k := f.k
	if k > 30 {
		return true
	}
	nBits := uint32(8 * (f.Len() - 1))
	delta := h>>17 | h<<15 //打乱hash值
	for j := uint8(0); j < k; j++ {
		bitPos := h % nBits
		f.bitmap[bitPos/8] |= 1 << (bitPos % 8)
		h += delta
	}
	return true
}

/**
返回 位图长度
*/

func (f *BloomFilter) Len() int32 {
	return int32(len(f.bitmap))
}

func (f *BloomFilter) MayContainKey(k []byte) bool {
	return f.MayContain(Hash(k))
}

func (f *BloomFilter) MayContain(h uint32) bool {
	if f.Len() < 2 {
		return false
	}
	k := f.k
	if k > 30 {
		return true
	}
	nBits := uint32(8 * (f.Len() - 1))
	delta := h>>17 | h<<15
	for j := uint8(0); j < k; j++ {
		bitPos := h % nBits
		if f.bitmap[bitPos/8]&(1<<(bitPos%8)) == 0 {
			return false
		}
		h += delta
	}
	return true
}

/**
如果h在bf里面就返回true
不在bf里面就返回false并插入进去
*/
func (f *BloomFilter) Allow(h uint32) bool {
	if f == nil {
		return true
	}
	already := f.MayContain(h)
	if !already {
		f.Insert(h)
	}
	return already

}

func Hash(b []byte) uint32 {
	const (
		seed = 0xbc9f1d34
		m    = 0xc6a4a793
	)
	h := uint32(seed) ^ uint32(len(b))*m
	for ; len(b) >= 4; b = b[4:] {
		h += uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
		h *= m
		h ^= h >> 16
	}
	switch len(b) {
	case 3:
		h += uint32(b[2]) << 16
		fallthrough
	case 2:
		h += uint32(b[1]) << 8
		fallthrough
	case 1:
		h += uint32(b[0])
		h *= m
		h ^= h >> 24
	}
	return h

}

func (b *BloomFilter) reset() {
	if b == nil {
		return
	}
	for i := range b.bitmap {
		b.bitmap[i] = 0
	}
}
