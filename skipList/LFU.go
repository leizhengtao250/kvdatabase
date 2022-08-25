package skiplist

import "container/list"

/**
  lru 对高频次的数据反而有不命中的问题
  lfu 难以应对稀疏的数据 存在旧数据长期不淘汰 ，要增加额外的空间记录和更新频率
**/

type lfu struct {
	cap   int
	list  *list.List
	cache map[string]uint32
	freq  int //访问频次
}
