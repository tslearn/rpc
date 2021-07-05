// +build windows

package metrics

import (
	"syscall"
	"unsafe"
)

var (
	kernel32DLL        = syscall.NewLazyDLL("kernel32")
	procGetSystemTimes = kernel32DLL.NewProc("GetSystemTimes")
)

type fileTime struct {
	DwLowDateTime  uint32
	DwHighDateTime uint32
}

func allCPUTimes() *cpuTimesStat {
	var lpIdleTime fileTime
	var lpKernelTime fileTime
	var lpUserTime fileTime
	r, _, _ := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&lpIdleTime)),
		uintptr(unsafe.Pointer(&lpKernelTime)),
		uintptr(unsafe.Pointer(&lpUserTime)),
	)
	if checkEqual(int(r), 0) {
		return nil
	}

	LOT := 0.0000001
	HIT := LOT * 4294967296.0
	idle := (HIT * float64(lpIdleTime.DwHighDateTime)) +
		(LOT * float64(lpIdleTime.DwLowDateTime))
	user := (HIT * float64(lpUserTime.DwHighDateTime)) +
		(LOT * float64(lpUserTime.DwLowDateTime))
	kernel := (HIT * float64(lpKernelTime.DwHighDateTime)) +
		(LOT * float64(lpKernelTime.DwLowDateTime))
	system := kernel - idle

	return &cpuTimesStat{
		Idle:   idle,
		User:   user,
		System: system,
	}
}
