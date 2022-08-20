package skiplist

import (
	"bytes"
	"fmt"
	"sync/atomic"
)

type skipList struct {
	headOffset uint32
	arena      Arena
	height     int32
}

/**
返回插入的位置在level层的before和next之间
*/

func (s *skipList) findSpliceForLevel(key []byte, before uint32, level int) (uint32, uint32) {
	for {
		beforeNode := s.arena.GetNode(before)
		next := beforeNode.getNextOffset(level)
		nextNode := s.arena.GetNode(next)
		if nextNode == nil {
			return before, next
		}
		nextKey := nextNode.key(s.arena)
		cmp := CompareKeys(beforeNode.key(s.arena), nextNode.key(s.arena))

	}

}

func CompareKeys(key1, key2 []byte) int {
	ConPanic((len(key1) <= 8 || len(key2) <= 8), fmt.Errorf("%s,%s < 8", string(key1), string(key2)))
	return bytes.Compare(key1[len(key1)-8:], key2[len(key2)-8:])
}

func ConPanic(condition bool, err error) {
	if condition {
		panic(err)
	}
}

func (s *skipList) findNear(key []byte, less bool, allowEqual bool) (*Node, bool) {
	x := s.getHead()
	level := int(s.getHeight() - 1)
	for {
		// Assume x.key < key.
		next := s.getNext(x, level)
		if next == nil {
			// x.key < key < END OF LIST
			if level > 0 {
				// Can descend further to iterate closer to the end.
				level--
				continue
			}
			// Level=0. Cannot descend further. Let's return something that makes sense.
			if !less {
				return nil, false
			}
			// Try to return x. Make sure it is not a head node.
			if x == s.getHead() {
				return nil, false
			}
			return x, false
		}

		nextKey := next.key(s.arena)
		cmp := CompareKeys(key, nextKey)
		if cmp > 0 {
			// x.key < next.key < key. We can continue to move right.
			x = next
			continue
		}
		if cmp == 0 {
			// x.key < key == next.key.
			if allowEqual {
				return next, true
			}
			if !less {
				// We want >, so go to base level to grab the next bigger note.
				return s.getNext(next, 0), false
			}
			// We want <. If not base level, we should go closer in the next level.
			if level > 0 {
				level--
				continue
			}
			// On base level. Return x.
			if x == s.getHead() {
				return nil, false
			}
			return x, false
		}
		// cmp < 0. In other words, x.key < key < next.
		if level > 0 {
			level--
			continue
		}
		// At base level. Need to return something.
		if !less {
			return next, false
		}
		// Try to return x. Make sure it is not a head node.
		if x == s.getHead() {
			return nil, false
		}
		return x, false
	}
}

func (s *skipList) getHead() *Node {
	return s.arena.GetNode(s.headOffset)
}

func (s *skipList) getNext(n *Node, height int) *Node {
	return s.arena.GetNode(n.getNextOffset(height))

}

func (s *skipList) Add(e *Entry) {
	key, v := e.key, ValueStruct{}
	listHeight := s.getHeight()

	var prev [maxHeight + 1]uint32
	var next [maxHeight + 1]uint32
	prev[listHeight] = s.headOffset             //最高层
	for i := int(listHeight) - 1; i >= 0; i-- { //开始从最高层往下层遍历
		prev[i], next[i] = s.findSpliceForLevel(key, prev[i+1], i)//找到第i层插入的位置
		if prev[i]==next[i]{
			v:=
		}

	}

}

func (s *skipList) getHeight() int32 {
	return atomic.LoadInt32(&s.height)
}

func (s *skipList) Get() {

}
