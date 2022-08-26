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
	加入cmrow数组存的bit数组是
	0:0111 0101
	1:1000 0001
	2:1100 1010

	这个里面存了6个counter
	有3个bit数组
	如果hash之后的值为n=5
	n/2 = 2
	r[2]=1100 1010
	n%2*4=4
	那么就是bit位序号为4，由于计算序号是从左到右执行
	其实需要的就是1100这个数
	因此 1 1 0 0 1 0 1 0
         7 6 5 4 3 2 1 0

	r[2]>>4 = 0000 1100
			& 0000 1111
			  0000 1100
           ==1100 这个就是我们要取的数
*/

func (r cmRow) get(n uint64) byte {
	return byte(r[n/2]>>((n%2)*4)) & 0x0f
}

/**
一个counter代表一个计数器
若n=5
*/

func (r cmRow) increment(n uint64) {
	index := n / 2                 //2            2
	offset := (n % 2) * 4          //4            0
	v := (r[index>>offset]) & 0x0f //1100         0000 1010
	if v < 15 {
		/**
			1左移动4位就使得 4-7位序列号的值加1
			1左移动0位就使得0-3位序列号的值加1
		**/
		r[index] += (1 << offset)
	}
}

//cmRow 计数减半
func (r cmRow) reset() {
	for i := range r {
		r[i] = (r[i] >> 1) & 0x77
	}
}

//cmRow 清空
func (r cmRow) clear() {
	for i := range r {
		r[i] = 0
	}
}

/*
TODO
	快速计算接近新的二次幂的算法
	x=5  最接近且大于5的2次幂是8
	x=111    128=2^7
	x=25     32 =2^5

*/
func next2Power(x int64) int64 {
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x |= x >> 32
	x++
	return x

}

const cmDepth = 4

type cmSketch struct {
	rows [cmDepth]cmRow  //给缓存计数
	seed [cmDepth]uint64 //hash函数有关
	mask uint64          //TODO
}

func newCmSketch(numCounters int64) *cmSketch {
	if numCounters == 0 {
		panic("cmSketch:invalid numCounters")
	}
	//numCounters 一定是2次幂，1后面跟n个0
	numCounters = next2Power(numCounters)
	//mask相当于numsCounters取反
	sketch := &cmSketch{
		mask: uint64(numCounters - 1),
	}

}
