//go:build !windows && !plan9 && !linux && !openbsd
// +build !windows,!plan9,!linux,!openbsd

package bbolt

import "os"

// fdatasync flushes written data to a file descriptor.
func fdatasync(file *os.File) error {
	return file.Sync()
}
