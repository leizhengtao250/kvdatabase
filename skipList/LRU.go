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

type storeItem struct {
	stage    int
	key      uint64
	conflict uint64
	value    interface{}
}

func newWindowLRU(size int, data map[uint64]*list.Element) *windowLRU {
	return &windowLRU{
		data: data,
		cap:  size,
		list: list.New(),
	}
}

func (lru *windowLRU) add(newitem storeItem) (eitem storeItem, evicted bool) {
	if lru.list.Len() < lru.cap {
		lru.data[newitem.key] = lru.list.PushFront(&newitem)
		return storeItem{}, false
	}
	//如果缓存容量是满的
	evictItem := lru.list.Back() //链表最后末尾的元素

	item := evictItem.Value.(*storeItem) //将元素转化为需要的类型
	//删除数据
	delete(lru.data, item.key)
	//这里对evictItem和*item赋值，避免向runtime再次申请空间
	eitem, *item = *item, newitem
	lru.data[item.key] = evictItem
	lru.list.MoveToFront(evictItem)
	return eitem, true
}
