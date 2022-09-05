package utils

import (
	"errors"
	"fmt"
	"hash/crc32"
	"path"
	"strconv"
	"strings"
)

var CastagnolicCrcTable = crc32.MakeTable(crc32.Castagnoli)

func CalculateCheckSum(data []byte) uint64 {
	return uint64(crc32.Checksum(data, CastagnolicCrcTable))
}

/**
取出文件编号
*/

func FID(name string) uint64 {
	/**
	a := "/opt/1.txt"
	a = path.Base(a)
	a为1.txt
	*/
	name = path.Base(name)
	//如果名称中没有.sst 那么返回0
	if !strings.HasSuffix(name, ".sst") {
		return 0
	}
	name = strings.TrimSuffix(name, ".sst") //12345.sst ->12345
	id, err := strconv.Atoi(name)
	if err != nil {
		Err(err)
		return 0
	}
	return uint64(id)

}

/**
data:需要计算的切片
expected：已经计算好的checkSum
*/

func VerifyChecksum(data []byte, expected []byte) error {
	actual := uint64(CalculateCheckSum(data))
	expectedU64 := ByteToU64(expected)
	if actual != expectedU64 {
		return errors.New(fmt.Sprintf("checksum mismatch ,actual: %d, expected: %d", actual, expectedU64))
	}
	return nil
}
