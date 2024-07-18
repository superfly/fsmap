//go:build !linux

package fsmap

type Entry struct {
	Physical uint64
	Length   uint64
}

func GetFreeBlocks(fs *os.File) (entries []Entry, err error) {
	return []Entry{}, nil
}
