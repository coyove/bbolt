//go:build js
// +build js

package bbolt

import (
	"os"
	"time"
)

// flock acquires an advisory lock on a file descriptor.
func flock(file *os.File, exclusive bool, timeout time.Duration) error {
	return nil
}

// funlock releases an advisory lock on a file descriptor.
func funlock(file *os.File) error {
	return nil
}

// mmap memory maps a DB's data file.
func mmap(file *os.File, sz int) ([]byte, error) {
	return nil, nil
}

// munmap unmaps a DB's data file from memory.
func munmap(data []byte) error {
	return nil
}
