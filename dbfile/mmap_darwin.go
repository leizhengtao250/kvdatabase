package dbfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type MmapFile struct {
	Data []byte
	Fd   *os.File
}

func OpenMmapFile(filename string, flag int, maxSz int) (*MmapFile, error) {
	//0666表示：创建了一个普通文件，所有人拥有对该文件的读、写权限，但是都不可执行
	fd, err := os.OpenFile(filename, flag, 0666)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("unable to open:%s ,error:%s", filename, err))
	}
	writeable := true
	if flag == os.O_RDONLY {
		writeable = false
	}
	return OpenMmapFileUsing(fd, maxSz, writeable)
}

func OpenMmapFileUsing(fd *os.File, sz int, writeable bool) (*MmapFile, error) {
	fileName := fd.Name()
	fi, err := fd.Stat()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot stat file:%s", fileName))
	}
	var rerr error
	fileSize := fi.Size()
	if sz > 0 && fileSize == 0 {
		//if file is empty ,truncate it to sz,not change IO offset
		if err := fd.Truncate(int64(sz)); err != nil {
			return nil, errors.New(fmt.Sprintf("error while truncate:%s", err))
		}
		fileSize = int64(sz)
	}
	buf, err := Mmap(fd, writeable, fileSize)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("while mmaping %s with size:%d", fileName, fileSize))
	}
	if fileSize == 0 {
		dir, _ := filepath.Split(fileName)
		go SyncDir(dir)

	}
	return &MmapFile{
		Data: buf,
		Fd:   fd,
	}, rerr

}

func SyncDir(dir string) error {
	df, err := os.Open(dir)
	if err != nil {
		return errors.New(fmt.Sprintf("while opening %s, error: %s", dir, err))
	}
	if err := df.Sync(); err != nil {
		return errors.New(fmt.Sprintf("while syncing %s,errors is %s", dir, err))
	}
	if err := df.Close(); err != nil {
		return errors.New(fmt.Sprintf("while closing %s,errors is %s", dir, err))
	}
	return nil
}
