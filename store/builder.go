package store

import (
	"errors"
	skiplist "kvdatabase/skipList"
	"kvdatabase/utils"
)

type tableBuilder struct {
	sstSize    int64
	curBlock   *block
	estimateSz int64
	blockList  []*block
	keycount   uint32
	opt *Options
}

type block struct {
	offset            int      //当前block的offset首地址
	checksum          []byte   //数据校验码
	entriesIndexStart int      //真正数据的起始位置
	chkLen            int      //数据校验码的起始位置
	data              []byte   //存放entryOffset,entryOffset的长度
	baseKey           []byte   //提取处baseKey节省空间
	entryOffsets      []uint32 //每条entry的起始位置
	end               int      //data[0:end]都是存放entryOffset,entryOffset的长度
	estimateSz        int64    //block创建时间
}

type buildData struct {
	blockList []*block
	index     []byte
	checksum  []byte
	size      int
}

type header struct {
	overlap uint16
	diff    uint16
}

func (tb *tableBuilder) done() buildData {
	tb.finishBlock()
	if len(tb.blockList) == 0 {
		return buildData{}
	}
	bd := buildData{
		blockList: tb.blockList,
	}
	var f skiplist.Filter
	if tb.
}

/*

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
	tb.curBlock = nil //当前block被序列化内存
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
