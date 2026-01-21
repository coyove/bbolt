package bbolt

import (
	"fmt"
	"os"
	"time"
	"unsafe"
)

type OsFile struct {
	file *os.File
	// `dataref` isn't used at all on Windows, and the golangci-lint
	// always fails on Windows platform.
	//nolint
	dataref []byte // mmap'ed readonly, write throws SEGV
	data    *[maxMapSize]byte
	datasz  int
}

func (f *OsFile) WriteAt(b []byte, off int64) (int, error) {
	return f.file.WriteAt(b, off)
}

func (f *OsFile) ReadAt(b []byte, off int64) (int, error) {
	return f.file.ReadAt(b, off)
}

func (f *OsFile) Name() string {
	return f.file.Name()
}

func (f *OsFile) Close() error {
	return f.file.Close()
}

func (f *OsFile) Fdatasync() error {
	return fdatasync(f.file)
}

func (f *OsFile) Truncate(size int64, _ bool) error {
	return f.file.Truncate(size)
}

func (f *OsFile) Size() (int64, error) {
	fi, err := f.file.Stat()
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

func (f *OsFile) Mmap(sz int) error {
	b, err := mmap(f.file, sz)
	if err != nil {
		return err
	}
	f.dataref = b
	f.data = (*[maxMapSize]byte)(unsafe.Pointer(&b[0]))
	f.datasz = sz
	return nil
}

func (f *OsFile) Minvalidate() {
	f.dataref = nil
	f.data = nil
	f.datasz = 0
}

func (f *OsFile) Munmap() error {
	err := munmap(f.dataref)
	f.Minvalidate()
	return err
}

func (f *OsFile) Flock(exclusive bool, timeout time.Duration) error {
	return flock(f.file, exclusive, timeout)
}

func (f *OsFile) Funlock() error {
	return funlock(f.file)
}

func (f *OsFile) Mlock(fileSize int) error {
	sizeToLock := fileSize
	if sizeToLock > f.datasz {
		// Can't lock more than mmaped slice
		sizeToLock = f.datasz
	}
	return mlock(f.dataref[:sizeToLock])
}

func (f *OsFile) Munlock(fileSize int) error {
	if f.dataref == nil {
		return nil
	}
	sizeToUnlock := fileSize
	if sizeToUnlock > f.datasz {
		// Can't unlock more than mmaped slice
		sizeToUnlock = f.datasz
	}
	return munlock(f.dataref[:sizeToUnlock])
}

func (f *OsFile) Mdata() *[maxMapSize]byte {
	return f.data
}

func (f *OsFile) MdataSize() int {
	return f.datasz
}

type MemFile struct {
	filedata []byte
	dataref  []byte
	data     *[maxMapSize]byte
	datasz   int
}

func (f *MemFile) WriteAt(b []byte, off int64) (int, error) {
	if int(off)+len(b) > len(f.filedata) {
		f.Truncate(off+int64(len(b)), false)
	}
	return copy(f.filedata[off:], b), nil
}

func (f *MemFile) ReadAt(b []byte, off int64) (int, error) {
	return copy(b, f.filedata[off:]), nil
}

func (f *MemFile) Name() string {
	return "<memory>"
}

func (f *MemFile) Close() error {
	return nil
}

func (f *MemFile) Fdatasync() error {
	return nil
}

func (f *MemFile) Truncate(size int64, forceResizing bool) error {
	if int(size) > cap(f.filedata) {
		if forceResizing {
		} else if len(f.filedata) > 0 {
			panic(fmt.Sprintf("truncate grow after db.init(): %d -> %d", len(f.filedata), size))
		}
		tmp := make([]byte, size)
		copy(tmp, f.filedata)
		f.filedata = tmp
	}
	f.filedata = f.filedata[:size]
	f.dataref = f.filedata
	f.data = (*[maxMapSize]byte)(unsafe.Pointer(&f.filedata[0]))
	return nil
}

func (f *MemFile) Size() (int64, error) {
	return int64(len(f.filedata)), nil
}

func (f *MemFile) Mmap(sz int) error {
	if sz > len(f.filedata) {
		tmp := make([]byte, sz)
		copy(tmp, f.filedata)
		f.filedata = tmp[:len(f.filedata)]
	}
	f.dataref = f.filedata
	f.data = (*[maxMapSize]byte)(unsafe.Pointer(&f.filedata[0]))
	f.datasz = sz
	return nil
}

func (f *MemFile) Minvalidate() {
	f.dataref = nil
	f.data = nil
	f.datasz = 0
}

func (f *MemFile) Munmap() error { f.Minvalidate(); return nil }

func (f *MemFile) Flock(exclusive bool, timeout time.Duration) error { return nil }

func (f *MemFile) Funlock() error { return nil }

func (f *MemFile) Mlock(fileSize int) error { return nil }

func (f *MemFile) Munlock(fileSize int) error { return nil }

func (f *MemFile) Mdata() *[maxMapSize]byte { return f.data }

func (f *MemFile) MdataSize() int { return f.datasz }
