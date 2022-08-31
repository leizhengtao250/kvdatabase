package skiplist

import "container/list"

/**
  lru 对高频次的数据反而有不命中的问题
  lfu 难以应对稀疏的数据 存在旧数据长期不淘汰 ，要增加额外的空间记录和更新频率
**/

type windowLRU struct {
	data map[uint64]*list.Element
	cap  int
	list *list.List
}

//存在缓存中的封装数据格式
type storeItem struct {
	stage    int //标注是在主缓存还是备缓存区
	key      uint64
	conflict uint64 //key如果出现冲突

	value interface{}
}

func newWindowLRU(size int, data map[uint64]*list.Element) *windowLRU {
	return &windowLRU{
		data: data,
		cap:  size,
		list: list.New(),
	}
}

/**
向windows中增加一个LRU数据
没有满  返回false
满了 	返回true

*/
func (lru *windowLRU) add(newitem storeItem) (eitem storeItem, evicted bool) {
	//如果这个数据已经存在于缓存中，那就更新这个数据，并掉到头部
	if _, ok := lru.data[newitem.key]; ok {
		v := lru.data[newitem.key]
		v.Value = newitem
		lru.list.MoveToFront(v)
		return storeItem{}, false
	}
	if lru.cap > lru.list.Len() {
		lru.data[newitem.key] = lru.list.PushFront(&newitem)
		return storeItem{}, false
	}
	//如果容量已满 而且 缓存中不存在newitem数据
	evictItem := lru.list.Back()
	item := evictItem.Value.(*storeItem)
	//从list中删除
	lru.list.Remove(evictItem)
	//从map中删除末尾数据
	delete(lru.data, item.key)
	//向map添加newitem数据,并放到头部
	lru.data[newitem.key] = lru.list.PushFront(&newitem)
	eitem = *item
	return eitem, true //返回被淘汰的数据
}

func (lru *windowLRU) get(v *list.Element) {
	lru.list.MoveToFront(v)
}
