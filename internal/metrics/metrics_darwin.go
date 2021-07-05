// +build darwin
// +build cgo

package metrics

/*
#include <mach/mach_host.h>
*/
import "C"

import (
	"unsafe"
)

func allCPUTimes() *cpuTimesStat {
	var count C.mach_msg_type_number_t
	var cpuLoad C.host_cpu_load_info_data_t

	count = C.HOST_CPU_LOAD_INFO_COUNT

	status := C.host_statistics(C.host_t(C.mach_host_self()),
		C.HOST_CPU_LOAD_INFO,
		C.host_info_t(unsafe.Pointer(&cpuLoad)),
		&count)

	if !checkEqual(int(status), int(C.KERN_SUCCESS)) {
		return nil
	}

	return &cpuTimesStat{
		User:   float64(cpuLoad.cpu_ticks[C.CPU_STATE_USER]),
		System: float64(cpuLoad.cpu_ticks[C.CPU_STATE_SYSTEM]),
		Nice:   float64(cpuLoad.cpu_ticks[C.CPU_STATE_NICE]),
		Idle:   float64(cpuLoad.cpu_ticks[C.CPU_STATE_IDLE]),
	}
}
