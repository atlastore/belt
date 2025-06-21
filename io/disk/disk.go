package disk

import "syscall"

type DiskSpace struct {
	stat *syscall.Statfs_t
	total uint64
}

func NewDiskSpace(path string) *DiskSpace {
	var stat syscall.Statfs_t
	syscall.Statfs(path, &stat)
	return &DiskSpace{stat: &stat, total: 0}
}

func (ds *DiskSpace) Free() uint64 {
	return ds.stat.Bfree * uint64(ds.stat.Bsize)
}

func (ds *DiskSpace) Available() uint64 {
	return ds.stat.Bavail * uint64(ds.stat.Bsize)
}

func (ds *DiskSpace) Total() uint64 {
	if ds.total == 0 {
		tot := uint64(ds.stat.Blocks) * uint64(ds.stat.Bsize)
		ds.total = tot
		return tot
	}
	return ds.total
}

func (ds *DiskSpace) Used() uint64 {
	return ds.Total() - ds.Free()
}

func (ds *DiskSpace) Usage() float64 {
	return float64(ds.Used()) / float64(ds.Total())
}

