package store

import (
	"errors"
	"fmt"
	proto2 "github.com/golang/protobuf/proto"
	"io"
	"kvdatabase/dbfile"
	"kvdatabase/proto"
	"kvdatabase/utils"
	"os"
	"sync"
	"syscall"
	"time"
)

//sstable

type SSTable struct {
	lock *sync.RWMutex
	f    *dbfile.MmapFile
	//为了加速查找数据，每个sst文件都会有最大key和最小key
	maxKey         []byte
	minKey         []byte
	indexStart     int
	indexLen       int
	fid            uint64
	createAt       time.Time
	indexTables    *proto.TableIndex
	hasBloomFilter bool
}

/**
打开一个sstable，具备mmap，
*/

func OpenSStable(opt *dbfile.OptionsSST) *SSTable {
	omf, err := dbfile.OpenMmapFile(opt.FileName, os.O_CREATE|os.O_RDWR, opt.MaxSz)
	utils.Err(err)
	return &SSTable{
		f:    omf,
		fid:  opt.FID,
		lock: &sync.RWMutex{},
	}
}

func (ss *SSTable) Bytes(off, sz int) ([]byte, error) {
	return ss.f.Bytes(off, sz)
}

func (ss *SSTable) Init() error {
	var ko *proto.BlockOffset
	var err error
	if ko, err = ss.initTable(); err != nil {
		return err
	}
	//从文件中获取创建时间
	stat, _ := ss.f.Fd.Stat()
	statType := stat.Sys().(*syscall.Stat_t)
	ss.createAt = time.Unix(statType.Atimespec.Sec, statType.Atimespec.Nsec)
	//init min key
	keyBytes := ko.GetKey()
	minKey := make([]byte, len(keyBytes))
	copy(minKey, keyBytes)
	ss.minKey = minKey
	ss.maxKey = minKey
	return nil
}

/**
initTable()是为了从sst中读取需要的数据
sst文件结构如下

低位 kv_data
.	offset
.	offset_len
.   checksum
.   checksum_len
高位


*/

func (ss *SSTable) initTable() (bo *proto.BlockOffset, err error) {
	//Data里面存的就是一张sst表
	readPos := len(ss.f.Data)
	readPos -= 4
	buf := ss.readCheckError(readPos, 4) //得到checkSum_length
	checkSum_len := int(utils.ByteToU32(buf))

	readPos -= checkSum_len
	buf_checksum := ss.readCheckError(readPos, checkSum_len) //得到checksum

	readPos -= 4
	buf = ss.readCheckError(readPos, 4)
	ss.indexLen = int(utils.ByteToU32(buf)) //得到索引长度

	readPos -= ss.indexLen
	ss.indexStart = readPos                               //索引起始位置
	data := ss.readCheckError(ss.indexStart, ss.indexLen) //索引值
	//对索引计算hash值和checksum做比较
	if err := utils.VerifyChecksum(data, buf_checksum); err != nil {
		return nil, errors.New(fmt.Sprintf("failed to verify checksum for table: %s", ss.f.Fd.Name()))
	}
	indexTable := &proto.TableIndex{}
	//把b中信息解析到indexTable中
	if err := proto2.Unmarshal(data, indexTable); err != nil {
		return nil, err
	}
	ss.indexTables = indexTable
	ss.hasBloomFilter = len(indexTable.BloomFilter) > 0
	if len(indexTable.GetOffsets()) > 0 {
		return indexTable.GetOffsets()[0], nil
	}
	return nil, errors.New("read index fail, offset is nil")
}

/**
从mmapfile中的off位置开始读，读取sz大小的数据
*/

func (ss *SSTable) readCheckError(off, sz int) []byte {
	buf, err := ss.read(off, sz)
	utils.Panic(err)
	return buf
}

func (ss *SSTable) read(off, sz int) ([]byte, error) {
	if len(ss.f.Data) > 0 {
		if len(ss.f.Data[off:]) < sz { //sz > Data[off:]切片的length
			return nil, io.EOF
		}
		return ss.f.Data[off : off+sz], nil
	}
	res := make([]byte, sz)
	_, err := ss.f.Fd.ReadAt(res, int64(off))
	return res, err

}
