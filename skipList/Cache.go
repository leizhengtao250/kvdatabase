package skiplist

import (
	"container/list"
	"kvdatabase/utils"
	"unsafe"

	xxhash "github.com/cespare/xxhash/v2"
)

type Cache struct {
	lru       *windowLRU
	slru      *segmentedLRU
	door      *utils.BloomFilter
	c         *cmSketch
	t         int32 //总共的访问次数
	threshold int32 //一个次数访问次数上限，达到就减半
	data      map[uint64]*list.Element
}

func NewCache(size int) *Cache {
	//定义window-lru所占百分比1%
	const lruPct = 1
	//计算出window-lru部分的容量
	lruSize := (lruPct * size) / 100

	if lruSize < 1 {
		lruSize = 1
	}
	//计算LFU部分的缓存容量
	slruSz := int(float64(size) * (100 - lruPct) / 100)
	if slruSz < 1 {
		slruSz = 1
	}
	//第一阶段
	slru1 := int(0.2 * float64(slruSz))

	data := make(map[uint64]*list.Element, size)
	return &Cache{
		lru:  newWindowLRU(lruSize, data),
		slru: newSLRU(data, slru1, slruSz-slru1),
		door: newFilter(size, 0.01),
		c:    newCmSketch(int64(size)),
	}

}

type stringStruct struct {
	str unsafe.Pointer
	len int
}

//go:noescape
//go:linkname memhash runtime.memhash
func memhash(p unsafe.Pointer, h, s uintptr) uintptr

func MemHashString(str string) uint64 {
	ss := (*stringStruct)(unsafe.Pointer(&str))
	return uint64(memhash(ss.str, 0, uintptr(ss.len)))
}

func Memhash(data []byte) uint64 {
	ss := (*stringStruct)(unsafe.Pointer(&data))
	return uint64(memhash(ss.str, 0, uintptr(ss.len)))
}

func (c *Cache) keyToHash(key interface{}) (uint64, uint64) {
	if key == nil {
		return 0, 0
	}
	switch k := key.(type) {
	case uint64:
		return k, 0
	case string:
		return MemHashString(k), xxhash.Sum64String(k)
	case []byte:
		return Memhash(k), xxhash.Sum64(k)
	case byte:
		return uint64(k), 0
	case int:
		return uint64(k), 0
	case int32:
		return uint64(k), 0
	case uint32:
		return uint64(k), 0
	case int64:
		return uint64(k), 0
	default:
		panic("Key type not supported")
	}

}

func (c *Cache) Set(key, value interface{}) bool {
	if key == nil {
		return false
	}
	keyHash, confilctHash := c.keyToHash(key)
	//刚放进去的缓存都是先到window-lru,因此stage=0
	i := storeItem{
		stage:    0,
		key:      keyHash,
		value:    value,
		conflict: confilctHash,
	}
	//如果在插入新数据以后window-lru已经满了，要返回被淘汰的数据
	eitem, evicted := c.lru.add(i)
	if !evicted { //如果没有满，直接返回插入成功
		return true
	}

	//window-lru淘汰了部分数据
	victm := c.slru.victim() //这个是slru淘汰的
	//假如LFU没有满,直接把从win-LRU中淘汰的数据加入到stageOne中
	if victm == nil {
		c.slru.add(eitem)
		return true
	}

	/**
	这个地方是win-lru和slru都满了
	进行pk，比较访问次数 决定去留
	这个时候需要和stageOne部分找一个淘汰者
	再次确定window-lru淘汰的数据还存在于缓存中
	如果被其他协程移除了缓存，那么就没必要在继续pk
	直接插入window-lru即可
	*/
	if !c.door.Allow(uint32(eitem.key)) {
		return true
	}
	//估算windowLru和LFU中淘汰数据，历史访问频次
	//访问频率高的，被认为更有资格留下来
	wcount := c.c.Estimate(eitem.key)  //win-lru要淘汰数据访问频数
	s1count := c.c.Estimate(victm.key) //stageOne中要淘汰的数据访问频数
	if wcount < s1count {              //win-lru中淘汰数据应该被淘汰，保留s1中的缓存数据
		return true
	}

	//wcount>s1count 从win-lru淘汰的数据要放入s1中，因此s1中要淘汰一个
	c.slru.add(eitem)
	return true

}

/**

 */
func (c *Cache) get(key uint64) (interface{}, bool) {
	c.t++                   //访问次数加1
	if c.t == c.threshold { //如果访问次数达到上限
		c.c.Reset()
		c.door.reset() //位图清0
		c.t = 0
	}
	val, ok := c.data[key]
	//如果key不存在
	if !ok {
		c.c.Increment(key)
		return nil, false
	}
	item := val.Value.(*storeItem)
	c.c.Increment(key) //在key的位图上加1

	v := item.value //取出value值

	if item.stage == 0 { //如果在0阶段，就是在window-lru中
		c.lru.get(val) //对缓存和链表做调整
	} else { //如果在slru中
		c.slru.get(val) //对缓存和链表做调整
	}
	return v, true
}
