package utils

import "hash/crc32"

var CastagnolicCrcTable = crc32.MakeTable(crc32.Castagnoli)

func CalculateCheckSum(data []byte) uint64 {
	return uint64(crc32.Checksum(data, CastagnolicCrcTable))
}
