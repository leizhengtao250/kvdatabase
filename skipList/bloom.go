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

}
