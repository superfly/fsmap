# fsmap

Go module to list free block ranges on Linux filesystems by calling the [`GETFSMAP` ioctl](https://man7.org/linux/man-pages/man2/ioctl_getfsmap.2.html).
Works with a raw block device containing a mountable ext4 filesystem (which it will temporarily mount), or an already-mounted filesystem.

## Example

```shell
$ go get github.com/superfly/fsmap
```

```go
package main

import "github.com/superfly/fsmap"
import "os"

func main() {
	file, _ := os.Open(os.Args[1])
	entries, _ := fsmap.GetFreeBlocks(file)
	free, total := 0, 0
	for _, entry := range entries {
		println(entry.Physical, "(", entry.Length, "bytes)")
		free += int(entry.Length)
		total = int(entry.Physical + entry.Length)
	}
	println("Total free:", int(float64(free)/float64(total)*100),"%")
}
```

```shell
$ fallocate -l 1G file &&
  mkfs.ext4 -qF file &&
  mkdir -p mount &&
  sudo mount -ro loop file mount

# mountpoint argument
$ sudo ./fsmap-test mount
17395712 ( 116822016 bytes)
134746112 ( 267907072 bytes)
403181568 ( 133689344 bytes)
570425344 ( 100663296 bytes)
671617024 ( 267907072 bytes)
940052480 ( 133689344 bytes)
Total free: 95 %

# block device argument
$ sudo ./fsmap-test $(losetup -j file | cut -d: -f1)
Free space at: 17395712 116822016
Free space at: 134746112 267907072
Free space at: 403181568 133689344
Free space at: 570425344 100663296
Free space at: 671617024 267907072
Free space at: 940052480 133689344
Total free: 95 %
```
