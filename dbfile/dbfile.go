package dbfile

type OptionsSST struct {
	FID      uint64
	FileName string //文件名称
	Dir      string //文件目录
	Path     string //文件路径
	Flag     int    //文件权限
	MaxSz    int    //最大占用空间的大小
}
