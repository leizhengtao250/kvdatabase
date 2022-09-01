package dbfile

import (
	"os"

	"golang.org/x/sys/unix"
)

func mmap(fd *os.File, writeable bool, size int64) ([]byte, error) {
	mtype := unix.PROT_READ
	if writeable {
		mtype |= unix.PROT_WRITE
	}
	return unix.Mmap(int(fd.Fd()), 0, int(size), mtype, unix.MAP_SHARED)
}
