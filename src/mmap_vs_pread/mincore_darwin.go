//go:build darwin

package main

import (
	"syscall"
	"unsafe"
)

// x/sys/unix does not wrap mincore(2) for darwin, but the syscall exists
// (SYS_MINCORE = 78). Same contract as Linux: one result byte per page,
// bit 0 means "resident, referencing it will not hit disk".
func mincoreVec(b []byte, vec []byte) error {
	_, _, errno := syscall.Syscall(78,
		uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), uintptr(unsafe.Pointer(&vec[0])))
	if errno != 0 {
		return errno
	}
	return nil
}
