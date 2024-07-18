package fsmap

//go:generate sh -c "go tool cgo -godefs _fsmap_defs_linux.go > fsmap_defs_linux.go && rm -rf _obj"

import (
	"fmt"
	"log"
	"math"
	"os"
	"syscall"
	"unsafe"
)

type Entry struct {
	Physical uint64
	Length   uint64
}

// temporarily mounts an ext4 block device read-only, calls a function with the mount path, and unmounts the device.
func withMount(device *os.File, mountFunc func(fs *os.File)) error {
	temp, err := os.MkdirTemp(os.TempDir(), "mount")
	if err != nil {
		return fmt.Errorf("failed to create temp mount path: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(temp); err != nil {
			log.Printf("failed to remove temp mount path: %v", err)
		}
	}()

	err = syscall.Mount(device.Name(), temp, "ext4", syscall.MS_RDONLY, "")
	if err != nil {
		return fmt.Errorf("mount failed: %w", err)
	}
	defer func() {
		if err := syscall.Unmount(temp, 0); err != nil {
			log.Printf("unmount failed: %v", err)
		}
	}()

	mountFile, err := os.OpenFile(temp, os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("failed to open mount file: %w", err)
	}
	defer func() {
		if err := mountFile.Close(); err != nil {
			log.Printf("failed to close mount file: %v", err)
		}
	}()

	mountFunc(mountFile)
	return nil
}

// GetFreeBlocks returns a set of free-block entries for a filesystem.
// The fs argument may be either a raw block device containing a mountable filesystem,
// or an already-mounted filesystem.
func GetFreeBlocks(fs *os.File) (entries []Entry, err error) {
	getEntries := func(fs *os.File) { entries, err = getFreeBlocks(fs) }
	info, err := fs.Stat()
	if err != nil {
		return
	}
	if info.Mode().Type() == os.ModeDevice {
		if err2 := withMount(fs, getEntries); err2 != nil {
			err = err2
		}
	} else {
		getEntries(fs)
	}
	if err != nil {
		err = fmt.Errorf("failed to get free blocks: %w", err)
	}
	return
}

var highKey = FSMap{math.MaxUint32, math.MaxUint32, math.MaxUint64, math.MaxUint64, math.MaxUint64, 0, [3]uint64{}}

// calls FS_IOC_GETFSMAP to return a set of free-block entries for the mounted filesystem.
func getFreeBlocks(fs *os.File) ([]Entry, error) {
	head := (*FSMapHead)(unsafe.Pointer(&make([]byte, Sizeof_FSMapEntries)[0]))
	head.Count = FSMapEntries
	head.Keys[1] = highKey
	freeEntries := make([]Entry, 0)
	for last := false; last == false; {
		headPtr := unsafe.Pointer(head)
		if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fs.Fd(), FS_IOC_GETFSMAP, uintptr(headPtr)); errno != 0 {
			return nil, errno
		}
		if head.Entries == 0 {
			break
		}
		entries := unsafe.Slice((*FSMap)(unsafe.Add(headPtr, Sizeof_FSMapHead)), head.Entries)
		for _, r := range entries {
			if r.Flags&FMR_OF_SPECIAL_OWNER != 0 && r.Owner == FMR_OWN_FREE {
				freeEntries = append(freeEntries, Entry{r.Physical, r.Length})
			}
			if r.Flags&FMR_OF_LAST != 0 {
				last = true
				break
			}
		}
		head.Keys[0] = entries[head.Entries-1]
	}

	return freeEntries, nil
}
