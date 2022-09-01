package store

/**
levelManger
管理整个level层级，其功能如下
1.从l0-l7层查询
2.给每层开辟空间，以放入数据
3.从内存接收数据flush到L0层
*/
type levelManager struct {
	maxFid uint64
	opt    *OptionLsm
}

/**
levelHandle
对具体层级做处理
1.增加一个batch
2.增加一个table
3.拓展空间
4.
*/
type levelHandler struct {
}

func newLevelManager(opt *){

}