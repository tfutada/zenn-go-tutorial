//go:build linux

package main

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

// Neither the stdlib syscall package nor x/sys/unix wraps mincore(2)
// for Linux, so call it by number. One result byte per page; bit 0
// means "resident, referencing it will not hit disk".
func mincoreVec(b []byte, vec []byte) error {
	_, _, errno := unix.Syscall(unix.SYS_MINCORE,
		uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), uintptr(unsafe.Pointer(&vec[0])))
	if errno != 0 {
		return errno
	}
	return nil
}
