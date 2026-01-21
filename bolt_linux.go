package bbolt

import (
	"os"
	"syscall"
)

// fdatasync flushes written data to a file descriptor.
func fdatasync(file *os.File) error {
	return syscall.Fdatasync(int(file.Fd()))
}
