package store

import (
	"errors"
	"kvdatabase/proto"
	"kvdatabase/utils"
	"math"
	"unsafe"
)

type tableBuilder struct {
	sstSize       int64
	curBlock      *block //正捏在手上的curBlock,其他block满了，只有这个block有操作空间
	estimateSz    int64
	blockList     []*block
	keycount      uint32
	opt           *OptionLsm
	keyHashs      []uint32
	maxVersion    uint64
	staleDataSize int //Todo
}

type block struct {
	offset            int      //当前block的offset首地址
	checksum          []byte   //数据校验码
	entriesIndexStart int      //所有entry数据的起始位置
	chkLen            int      //数据校验码的长度
	data              []byte   //存放entryOffset,entryOffset的长度
	baseKey           []byte   //提取处baseKey节省空间
	entryOffsets      []uint32 //每条entry的在data中的起始位置
	end               int      //data[0:end]表示data的长度，其实也是block的占用空间大小
	estimateSz        int64    //block创建时间
}

type buildData struct {
	blockList []*block
	index     []byte
	checksum  []byte
	size      int //sst 文件的大小
}

type header struct {
	overlap uint16 //重叠key的长度
	diff    uint16 //非重叠key的长度
}

/**
@isStale :是否新鲜

 block-->kv_data->header
				  diffKey
                  expiresAt      存入
	  			  value
*/
func (tb *tableBuilder) add(e *utils.Entry, isStale bool) {
	key := e.Key
	val := utils.ValueStruct{
		Meta:      e.Meta,
		Value:     e.Value,
		ExpiresAt: e.ExpiresAt,
	}
	//新加入一条entry
	//看是否需要分配一个新的block，也为新的entry的加入分配空间
	if tb.tryFinishBlock(e) {
		if isStale {
			tb.staleDataSize += len(key) + 4
		}
	}
	tb.finishBlock()
	tb.curBlock = &block{
		data: make([]byte, tb.opt.BlockSize),
	}
	//将key的8位之后的byte做hash，然后放入keyHashs中，
	tb.keyHashs = append(tb.keyHashs, utils.Hash(utils.ParseKey(key)))

	if version := utils.ParseTs(key); version > tb.maxVersion {
		tb.maxVersion = version
	}
	var diffKey []byte
	if len(tb.curBlock.baseKey) == 0 {
		//如果没有basekey，那么将当前key作为basekey
		tb.curBlock.baseKey = append(tb.curBlock.baseKey[:0], key...)
		diffKey = key
	} else {
		diffKey = tb.keyDiff(key)
	}
	//baseKey 和 diffKey 长度要符合要求
	utils.CondPanic(!(len(key)-len(diffKey) <= math.MaxInt16), errors.New("tableBuilder.add: baseKey <= math.MaxUint16"))
	utils.CondPanic(!(len(diffKey) <= math.MaxInt16), errors.New("tableBuilder.add:diffKey <= math.MaxUint16"))

	h := header{
		overlap: uint16(len(key) - len(diffKey)),
		diff:    uint16(len(diffKey)),
	}
	//end 将end这个数字保存在entryOffsets中 TODO 这个是为啥
	//这个是用来计算block的长度的
	tb.curBlock.entryOffsets = append(tb.curBlock.entryOffsets, uint32(tb.curBlock.end))
	tb.append(h.encode())
	tb.append(diffKey)
	//在block.data上分配了dst大小
	dst := tb.allocate(int(val.EncodeSize()))
	//将value存入data中
	val.EncodeValue(dst)
}

func (h header) encode() []byte {
	var b [4]byte
	*(*header)(unsafe.Pointer(&b[0])) = h
	return b[:]
}

/**
将newkey中的basekey剔除，然后返回剩下来的部分
*/
func (tb *tableBuilder) keyDiff(newKey []byte) []byte {
	var i int
	for i = 0; i < len(newKey) && i < len(tb.curBlock.baseKey); i++ {
		if newKey[i] != tb.curBlock.baseKey[i] {
			break
		}
	}
	return newKey[i:]
}

/**
检查是否需要再分配一个block
看这个entry加入进以后会不会超过block
返回true：需要再增加一个block
返回false:不需要增加
*/
func (tb *tableBuilder) tryFinishBlock(e *utils.Entry) bool {
	//如果curBlock不存在，那么就不需要flush了
	if tb.curBlock == nil {
		return true
	}
	if len(tb.curBlock.entryOffsets) <= 0 {
		return false
	}
	key := e.Key
	val := utils.ValueStruct{
		Meta:      e.Meta,
		Value:     e.Value,
		ExpiresAt: e.ExpiresAt,
	}
	/**
			|加入一个entry，会给block带来空间的增大
			|1.a    entryOffsets增加一个offset(uint32) 4B
			 1.b    增加checksum和其长度 8B+4B
			 1.c    data中增加一个offset的长度 4B
	       	这个数组会增加一个uint32，代表新加入的entry在block中的偏移
			|所以 tb.curBlock.entryOffsets的长度+1，每个uint32是4个字节，
	        |
	block--->2.entry带来的增大(kv_data)
	           增加一个header   4
	           增加一个diffKey  按照key的长度 len(key)
	           增加一个expires_at     val包含了expires_at
	           增加一个value   val.encodeSize
			|总的应该是：(x+1)*4+4+4+8+4
			|
	*/

	utils.CondPanic(!((uint32(len(tb.curBlock.entryOffsets))+1)*4+4+4+8+4 < math.MaxUint32), errors.New("Integer overflow"))
	//要增加的空间大小
	entriesOffsetSize := int64((len(tb.curBlock.entryOffsets)+1)*4 + //1.a    entryOffsets增加一个offset(uint32) 4B
		4 + //data中增加一个offset的长度 长度用uin32表示 4B
		8 + //sum64 int checksum checksum的占用空间
		4) //checksum length    checksum的长度

	//这个是预估总的空间的大小
	tb.curBlock.estimateSz = int64(tb.curBlock.end) + entriesOffsetSize + int64(4) + int64(len(key)) + int64(val.EncodeSize())
	//如果预估大小超出最大值
	utils.CondPanic(!(uint32(tb.curBlock.estimateSz) < math.MaxUint32), errors.New("Integer overflow"))
	return tb.curBlock.estimateSz > int64(tb.opt.BlockSize)
}

func (tb *tableBuilder) done() buildData {
	tb.finishBlock() //block都封存好了
	if len(tb.blockList) == 0 {
		return buildData{}
	}
	bd := buildData{
		blockList: tb.blockList,
	}
	var f utils.Filter
	if tb.opt.BloomFalsePositive > 0 {
		bits := utils.BitsPerkey(len(tb.keyHashs), tb.opt.BloomFalsePositive)
		f = utils.NewFilter1(tb.keyHashs, bits)
	}
	//构建sst文件索引
	index, dataSize := tb.buildIndex(f)
	checksum := tb.calculateCheckSum(index)
	bd.index = index
	bd.checksum = checksum
	bd.size = int(dataSize) + len(index) + len(checksum) + 4 + 4
	return bd
}

/**
建立一个sst中的索引，对key的数量，布隆过滤器的hash，sst版本号，每个block的偏移量

*/

func (tb *tableBuilder) buildIndex(bloom []byte) ([]byte, uint32) {
	tableIndex := &proto.TableIndex{}
	if len(bloom) > 0 {
		tableIndex.BloomFilter = bloom
	}
	tableIndex.KeyCount = tb.keycount
	tableIndex.MaxVersion = tb.maxVersion
	tableIndex.Offsets = tb.writeBlockOffsets()
	var dataSize uint32
	for i := range tb.blockList {
		dataSize += uint32(tb.blockList[i].end)
	}
	data, err := tableIndex.Marshal()
	utils.Panic(err)
	return data, dataSize
}

/**
整个sst中block索引都写入proto找那个
将block序列化，其offset数据都放到proto中
*/
func (tb *tableBuilder) writeBlockOffsets() []*proto.BlockOffset {
	var startOffset uint32 //0
	var offsets []*proto.BlockOffset
	for _, bl := range tb.blockList {
		offset := tb.writeBlockOffset(bl, startOffset)
		offsets = append(offsets, offset)
		startOffset += uint32(bl.end)
	}
	return offsets
}

/**
单个block索引
*/
func (tb *tableBuilder) writeBlockOffset(bl *block, startOffset uint32) *proto.BlockOffset {
	offset := &proto.BlockOffset{}
	offset.Key = bl.baseKey
	offset.Len = uint32(bl.end)
	offset.Offset = startOffset
	return offset
}

/*
	由于这个block要满了
	那么将这个block封存
	要将这个block中存的entryOffsets([]uint32)和entryOffsets的长度，整个存储的entry的checksum,以及长度
	把这些数据存放到data中，那么在从内存到磁盘中的时候就可以直接把data中的数据存到磁盘中就可以了
*/

func (tb *tableBuilder) finishBlock() {
	if tb.curBlock == nil || len(tb.curBlock.entryOffsets) == 0 {
		return //这个tableBuilder下面没有block
	}
	tb.append(utils.U32SliceToBytes(tb.curBlock.entryOffsets))
	tb.append(utils.U32TOBytes(uint32(len(tb.curBlock.entryOffsets))))
	checksum := tb.calculateCheckSum(tb.curBlock.data[:tb.curBlock.end])
	tb.append(checksum)
	tb.append(utils.U32TOBytes(uint32(len(checksum))))
	tb.estimateSz += tb.curBlock.estimateSz
	tb.blockList = append(tb.blockList, tb.curBlock)
	//TODO:预估builder写入磁盘以后，sst文件的大小
	tb.keycount += uint32(len(tb.curBlock.entryOffsets))
	tb.curBlock = nil //当前block里的数据都做好准备了，可以扩充下一个block了
	return
}

//把data数据追加到block的data中
func (tb *tableBuilder) append(data []byte) {
	dst := tb.allocate(len(data))
	utils.CondPanic(len(data) != copy(dst, data), errors.New("tableBuilder append data"))
}

/**
在bb.data[]上分配了need空间大小
*/
func (tb *tableBuilder) allocate(need int) []byte {
	bb := tb.curBlock
	if len(bb.data[bb.end:]) < need { //现有的比需求的低
		sz := 2 * len(bb.data)
		if bb.end+need > sz {
			sz = bb.end + need
		}
		tmp := make([]byte, sz)
		copy(tmp, bb.data)
		bb.data = tmp
	}
	bb.end += need
	return bb.data[bb.end-need : bb.end]
}

func (tb *tableBuilder) calculateCheckSum(data []byte) []byte {
	//计算其crc码
	checkSum := utils.CalculateCheckSum(data)
	return utils.U64TOBytes(checkSum)
}
