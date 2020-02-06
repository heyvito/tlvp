package mem

import "unsafe"

const MaxUintptr = ^uintptr(0)
const PtrSize = 4 << (^uintptr(0) >> 63) // unsafe.Sizeof(uintptr(0)) but an ideal const
const _64bit = 1 << (^uintptr(0) >> 63) / 2
const heapAddrBits = (_64bit*(1-goArchWasm)*(1-goOSAix))*48 + (1-_64bit+goArchWasm)*(32-(goArchMips+goArchMipsle)) + 60*goOSAix
const maxAlloc = (1 << heapAddrBits) - (1-_64bit)*1

// MulUintptr returns a * b and whether the multiplication overflowed.
// On supported platforms this is an intrinsic lowered by the compiler.
func MulUintptr(a, b uintptr) (uintptr, bool) {
	if a|b < 1<<(4*PtrSize) || a == 0 {
		return a * b, false
	}
	overflow := b > MaxUintptr/a
	return a * b, overflow
}

func IsValidSize(size uint64) bool {
	mem, overflow := MulUintptr(unsafe.Sizeof(byte(0)), uintptr(size))
	if overflow || mem > maxAlloc || size < 0 {
		return false
	}
	return true
}
