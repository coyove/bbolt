//go:build !windows && !plan9 && !solaris && !aix && !js
// +build !windows,!plan9,!solaris,!aix,!js

package bbolt

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// flock acquires an advisory lock on a file descriptor.
func flock(file *os.File, exclusive bool, timeout time.Duration) error {
	var t time.Time
	if timeout != 0 {
		t = time.Now()
	}
	fd := file.Fd()
	flag := syscall.LOCK_NB
	if exclusive {
		flag |= syscall.LOCK_EX
	} else {
		flag |= syscall.LOCK_SH
	}
	for {
		// Attempt to obtain an exclusive lock.
		err := syscall.Flock(int(fd), flag)
		if err == nil {
			return nil
		} else if err != syscall.EWOULDBLOCK {
			return err
		}

		// If we timed out then return an error.
		if timeout != 0 && time.Since(t) > timeout-flockRetryTimeout {
			return ErrTimeout
		}

		// Wait for a bit and try again.
		time.Sleep(flockRetryTimeout)
	}
}

// funlock releases an advisory lock on a file descriptor.
func funlock(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}

// mmap memory maps a DB's data file.
func mmap(file *os.File, sz int) ([]byte, error) {
	// Map the data file to memory.
	b, err := unix.Mmap(int(file.Fd()), 0, sz, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	// Advise the kernel that the mmap is accessed randomly.
	err = unix.Madvise(b, syscall.MADV_RANDOM)
	if err != nil && err != syscall.ENOSYS {
		// Ignore not implemented error in kernel because it still works.
		return nil, fmt.Errorf("madvise: %s", err)
	}

	// Save the original byte slice and convert to a byte array pointer.
	return b, nil
}

// munmap unmaps a DB's data file from memory.
func munmap(data []byte) error {
	if data == nil {
		return nil
	}

	// Unmap using the original byte slice.
	return unix.Munmap(data)
}
