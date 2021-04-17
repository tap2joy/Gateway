package test

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestPoint2String(t *testing.T) {
	var a uint = 0
	p := uintptr(unsafe.Pointer(&a))
	for i := 0; i < int(unsafe.Sizeof(a)); i++ {
		p += 1
		pb := (*byte)(unsafe.Pointer(p))
		*pb = 1
	}

	fmt.Printf("%x\n", a) //0x1010100
}
