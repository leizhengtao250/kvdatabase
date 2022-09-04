package skiplist

import (
	"bytes"
	"fmt"
	"kvdatabase/utils"
	"math"
	"sync/atomic"
)

const (
	heightIncrease = math.MaxUint32 / 3
	maxHeight      = 20
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
		beforeNode := s.arena.GetNode(before) //
		next := beforeNode.getNextOffset(level)
		nextNode := s.arena.GetNode(next)
		if nextNode == nil {
			return before, next
		}
		nextKey := nextNode.key(s.arena)
		cmp := CompareKeys(key, nextKey)
		if cmp == 0 {
			return next, next
		}
		if cmp < 0 {
			return before, next
		}
		before = next
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

/***
	返回key最近的节点，这些节点有可能比key大，有可能比key小
	如果less=true，返回的是小于key的节点
	如果less=false,返回的是大于key的节点
	如果allowEqual=true,返回的是允许等于key的节点

**/

func (s *skipList) findNear(key []byte, less bool, allowEqual bool) (*Node, bool) {
	x := s.getHead()
	level := int(s.getHeight() - 1)
	for {
		// Assume x.key < key.
		/**
			next是指在level层上x的next节点指针
		**/
		next := s.getNext(x, level)

		//如果搜索到level层的最右端为空
		//就减小层数接着搜索
		if next == nil {
			// x.key < key < END OF LIST
			if level > 0 {
				// Can descend further to iterate closer to the end.
				level--
				continue
			}
			/** 如果是最后一层，且找到末尾
				如果此时less是false，即要找>给定可以的最小值
				如果找到末尾，说明不存在比key大的值，只能返回nil
			**/
			if !less {
				return nil, false
			}
			// 判断一下，key是不是头节点
			//如果是头结点，说明skiplist当中无任何元素
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

func (s *skipList) getHeight() int32 {
	return atomic.LoadInt32(&s.height)
}

/**
	heightIncress = max的uint32/2
	那么FastRand()会随机的选择 uint32的值，那么就有50%的几率FastRand() <= heightIncrease
	因此每执行一次randomHeight() 几率为50%
	那么h=1  ->50%
	   h=2   ->25%
	   h=3   ->12.5%
       h=4   ->.....
**/
func (s *skipList) randomHeight() int {
	h := 1
	for h < maxHeight && FastRand() <= heightIncrease {
		h++
	}
	return h
}

/**
	一条数据过来，首先提取处key ,value
	根据key，value，和层数，已经所属的内存管理器 构建出一个节点
	因此需要newNode(arena,key,value,height)
	key,value,已经提取，剩下来分为两步
	1.插入到skipList表，需要知道这个节点处于哪个层级
	2.数据要插入到arena中

	对于1：
		1.首先确定好层级，也就是这个节点的最高层级，使用随机概率层级的height，那么该层级以下的
		当前节点都会存在
		2，在该层中寻找合适位置，插入
	对于2：
		arena管理的是offset和size，其他不知道
		skipList来管理每个节点的信息
	节点：new之后得到新节点node
		1.先将真实的key，value放入arena，得到keyoffst,keysize,valueoffset,valuesize
		2.将key，value的offset，size 放到node节点中
		3.在arena中开辟空间，放入node


**/
func (s *skipList) Add(e *Entry) {
	key, v := e.key, ValueStruct{}
	listHeight := s.getHeight()

	var prev [maxHeight + 1]uint32              //0-maxHeight
	var next [maxHeight + 1]uint32              //0-maxHeight
	prev[listHeight] = s.headOffset             //最高层
	for i := int(listHeight) - 1; i >= 0; i-- { //开始从最高层往下层遍历
		prev[i], next[i] = s.findSpliceForLevel(key, prev[i+1], i) //找到第i层插入的位置
		//这里假设是cmp==0时,插入的key和next的key相等
		//由于skipList中不允许存在两个相同key的节点
		//因此，只需要对相同key的节点做更新操作
		if prev[i] == next[i] {
			offsetV := s.arena.PutVal(v)                 //value的offset，同时也存储在arena上
			vs := encodedValue(offsetV, v.EncodedSize()) //skipList中节点上存的value值
			preNode := s.arena.GetNode(prev[i])
			preNode.setValue(&s.arena, vs)
			//skipList表中只需要改变一个节点即可，因为同个节点用tower指针指向不同层
			return
		}

	}

	//运行到这里说明key可以正常插入到表中，排除了更新操作
	//在height层上创建一个新的node节点,那么它的下层节点必存在，用一个循环即可
	height := s.randomHeight()
	//新建的节点只要建立一个即可
	//每层的节点指针依靠node中tower的数组去指定
	x := newNode(s.arena, key, v, height)
	listHeight = s.getHeight()

	//用cas的方法增加高度
	for height > int(listHeight) {
		if atomic.CompareAndSwapInt32(&s.height, listHeight, int32(height)) {
			break
		}
		listHeight = s.getHeight()
	}

	for i := 0; i < height; i++ {
		for {
			if s.arena.GetNode(prev[i]) == nil {
				AssertTrue(i > 1)
				//从第0层开始循环插入找到插入点
				prev[i], next[i] = s.findSpliceForLevel(key, s.headOffset, i)
				AssertTrue(prev[i] != next[i])
			}
			//key->next
			x.tower[i] = next[i]
			pnode := s.arena.GetNode(prev[i])
			//pnode.tower[i]=当前新建立节点的offset
			//也就是pre->key
			if pnode.casNextOffset(i, next[i], s.arena.getNodeOffset(x)) {
				break
			}

			/**
			如果pre->key的原子过程失败，要重新计算插入的位置
			因为产生了并发，修改了pre值，导致了pre[i]和x 相等要重新计算
			*/
			prev[i], next[i] = s.findSpliceForLevel(key, prev[i], i)

			if prev[i] == next[i] {
				AssertTruef(i == 0, "Equality can happen only on base level: %d", i)
				offsetV := s.arena.PutVal(v)                 //value的offset，同时也存储在arena上
				vs := encodedValue(offsetV, v.EncodedSize()) //skipList中节点上存的value值
				preNode := s.arena.GetNode(prev[i])
				preNode.setValue(&s.arena, vs)
				//skipList表中只需要改变一个节点即可，因为同个节点用tower指针指向不同层
				return
			}

		}
	}

}

/**
	skipList中value和arena中的value有区别
	1.skipList中value为32位offset，32位的size
	2.根据32位offset和32位的size，再去arena中寻找真正的value值
	3.返回的是组装的uint64
**/

func encodedValue(offset, size uint32) uint64 {
	return uint64(size)<<32 | uint64(offset) //小端存储
}

/**
通过前8位的比较，得到
*/
func (s *skipList) Search(key []byte) ValueStruct {
	n, _ := s.findNear(key, false, true)
	if n == nil {
		return ValueStruct{}
	}
	/**
	因为findNear 的less=false
	那么寻找的是节点>=key
	*/
	nextKey := s.arena.GetKey(n.keyOffset, n.keySize)
	//比较key和nextkey
	if !utils.SameKey(key, nextKey) {
		return ValueStruct{}
	}
	valOffset, valSize := n.getValueOffset()
	vs := s.arena.GetVal(valOffset, valSize)
	vs.ExpiresAt = utils.ParseTs(nextKey)
	return vs
}
