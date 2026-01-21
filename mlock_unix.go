//go:build !windows && !js
// +build !windows,!js

package bbolt

import "golang.org/x/sys/unix"

// mlock locks memory of db file
func mlock(data []byte) error {
	return unix.Mlock(data)
}

// munlock unlocks memory of db file
func munlock(data []byte) error {
	return unix.Munlock(data)
}
