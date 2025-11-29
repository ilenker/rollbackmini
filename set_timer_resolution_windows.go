package main

import (
	"fmt"
    "syscall"
    "unsafe"
	"os"
)

var (
    ntdll				= syscall.NewLazyDLL("ntdll.dll")
    procNtSetTimerRes	= ntdll.NewProc("NtSetTimerResolution")
)

func init() {

    // Desired resolution: 1ms (10000 in 100ns units)
    desired := uint32(5000)
    var current uint32

    // Call NtSetTimerResolution(desired, TRUE, &current)
	_, _, err := procNtSetTimerRes.Call(
        uintptr(desired),
        uintptr(1),
        uintptr(unsafe.Pointer(&current)),
    )

	errno := err.(syscall.Errno)

	if errno == syscall.Errno(0) {
		fmt.Printf("Windows-specific: NtSetTimerResolution(%d) (success)\n", desired)
	} else {
		fmt.Printf("Windows-specific: NtSetTimerResolution(%d) (failure)\n", desired)
		fmt.Println(err)
		os.Exit(1)
	}
}

