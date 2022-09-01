package dbfile

import "os"

func Mmap(fd *os.File, writeable bool, size int64) ([]byte, error) {
	return mmap(fd, writeable, size)
}
