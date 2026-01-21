//go:build windows || js

package bbolt

// mlock locks memory of db file
func mlock(_ []byte) error {
	panic("mlock is supported only on UNIX systems")
}

// munlock unlocks memory of db file
func munlock(_ []byte) error {
	panic("munlock is supported only on UNIX systems")
}
