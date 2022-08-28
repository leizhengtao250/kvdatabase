package skiplist

import "container/list"

const (
	STAGE_TWO = 2
	STAGE_ONE = 1
)

/**
先进入one，如果在one再被访问的时候，在进入two
one 20%
two 80%

*/
type segmentedLRU struct {
	data                     map[uint64]*list.Element
	stageOneCap, stageTwoCap int
	stageOne, stageTwo       *list.List
}

func newSLRU(data map[uint64]*list.Element, size1, size2 int) *segmentedLRU {
	return &segmentedLRU{
		data:        data,
		stageOneCap: size1,
		stageTwoCap: size2,
		stageOne:    list.New(),
		stageTwo:    list.New(),
	}
}

func (s *segmentedLRU) add(newitem storeItem) {
	newitem.stage = 1
	if s.stageOne.Len() < s.stageOneCap || s.stageTwo.Len() < s.stageTwoCap {
		s.data[newitem.key] = s.stageOne.PushFront(&newitem)
		return
	}
	e := s.stageOne.Back()
	item := e.Value.(*storeItem)
	//淘汰末尾
	delete(s.data, item.key)
	*item = newitem
	s.data[item.key] = s.stageOne.PushFront(&newitem)
	//淘汰链表中的元素
	s.stageOne.Remove(e)

}

/**
 TODO:修改
**/
func (s *segmentedLRU) get(v *list.Element) {
	item := v.Value.(*storeItem)
	//如果访问的缓存数据在第二阶段
	if item.stage == STAGE_TWO {
		s.stageTwo.MoveToFront(v)
		return
	}
	//如果缓存数据在第一阶段且第二阶段没有满
	//那么需要将数据从第一阶段调入第二阶段
	if s.stageTwo.Len() < s.stageTwoCap {
		s.stageOne.Remove(v)
		item.stage = STAGE_TWO
		s.data[item.key] = s.stageTwo.PushFront(item)
		return
	}

	//如果缓存在第一阶段且此时第二阶段也满了，需要先把第二阶段的缓存淘汰一个
	//第二阶段淘汰的数据会重新进入第一阶段
	e := s.stageTwo.Back()
	item2 := e.Value.(*storeItem)

	item2.stage = STAGE_ONE
	item.stage = STAGE_TWO

	s.data[item.key] = v
	s.stageOne.PushFront(e)
	s.stageTwo.PushFront(v)
}

func (s *segmentedLRU) Len() int {
	return s.stageTwo.Len() + s.stageOne.Len()
}

//window-lru淘汰了部分数据
//这个时候需要在stageOne部分找一个淘汰者
//需要pk 决定去留
func (s *segmentedLRU) victim() *storeItem {
	if s.Len() < s.stageOneCap+s.stageTwoCap {
		return nil
	}
	v := s.stageOne.Back()
	return v.Value.(*storeItem)
}
