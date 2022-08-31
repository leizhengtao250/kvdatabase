package store

import "os"

type MmapFile struct {
	Data []byte   //trans data
	Fd   *os.File //file address
}
