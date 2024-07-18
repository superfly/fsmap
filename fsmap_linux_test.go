package fsmap

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
)

func tempFile(t *testing.T, size int64) *os.File {
	temp, err := os.CreateTemp(t.TempDir(), "fsmap_test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(temp.Name()) })
	if err := temp.Truncate(size); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	return temp
}

const size = 1024 * 1024 * 1024 * 16

func TestGetFreeBlocks(t *testing.T) {
	dev := getBlockDevice(t)
	entries, err := GetFreeBlocks(dev)
	if err != nil {
		t.Fatal(err)
	}

	testEntries(t, entries)
}

func TestGetFreeBlocksMounted(t *testing.T) {
	temp, err := os.MkdirTemp(os.TempDir(), "mount")
	t.Cleanup(func() {
		if err := os.RemoveAll(temp); err != nil {
			t.Fatalf("failed to remove temp mount path: %v", err)
		}
	})
	dev := getBlockDevice(t)
	if err = syscall.Mount(dev.Name(), temp, "ext4", syscall.MS_RDONLY, ""); err != nil {
		t.Fatalf("mount failed: %v", err)
	}
	t.Cleanup(func() {
		if err := syscall.Unmount(temp, 0); err != nil {
			t.Fatalf("unmount failed: %v", err)
		}
	})

	mount, err := os.Open(temp)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err = mount.Close(); err != nil {
			t.Fatalf("failed to close volume: %v", err)
		}
	})

	entries, err := GetFreeBlocks(mount)
	if err != nil {
		t.Fatal(err)
	}

	testEntries(t, entries)
}

func testEntries(t *testing.T, entries []Entry) {
	var totalFree uint64
	for _, entry := range entries {
		totalFree += entry.Length
	}
	free := float64(totalFree) / size * 100
	t.Logf("%v Free entries: %v / %v = %.02f%%", len(entries), totalFree, size, free)
	if free < 95 || free > 99 {
		t.Fatalf("Expected 95-99%% free space, got %.02f%%", free)
	}
}

func getBlockDevice(t *testing.T) *os.File {
	file := tempFile(t, size)
	cmd := exec.Command("losetup", "-f", "--show", file.Name())
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}

	dev, err := os.Open(strings.TrimSpace(string(out)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cmd = exec.Command("losetup", "-d", dev.Name())
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		_ = cmd.Run()
	})

	cmd = exec.Command("mkfs.ext4", dev.Name())
	if out, err = cmd.CombinedOutput(); err != nil {
		t.Log(out)
		t.Fatal(err)
	}
	return dev
}
