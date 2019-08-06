package common

import (
	"time"

	"github.com/shirou/gopsutil/cpu"
)

var fnCpuPercent = cpu.Percent

// GetCPUTotalPercentage get current cpu average usage
func GetCPUTotalPercentage(duration time.Duration) (float64, RPCError) {
	cpuUsage, err := fnCpuPercent(duration, false)
	if len(cpuUsage) != 1 {
		return 0, NewRPCError("cpu api called error")
	}
	if err != nil {
		return 0, WrapSystemErrorWithDebug(err)
	}
	return cpuUsage[0], nil
}
