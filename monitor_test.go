package common

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func Test_GetCPUTotalPercentage(t *testing.T) {
	assert := NewAssert(t)

	// GetCPUUsage for 100 ms
	cpuUsage, err := GetCPUTotalPercentage(100 * time.Millisecond)
	assert(err).IsNil()
	assert(cpuUsage >= 0 && cpuUsage <= 100).IsTrue()

	// hoot fnCpuPercent api, return error array
	fnCpuPercent = func(time.Duration, bool) (float64s []float64, e error) {
		ret := make([]float64, 2)
		return ret, nil
	}
	cpuUsage, err = GetCPUTotalPercentage(100 * time.Millisecond)
	fmt.Println(cpuUsage, err.GetMessage())
	assert(cpuUsage).Equals(float64(0))
	assert(err.GetMessage()).Equals("cpu api called error")

	// hoot fnCpuPercent api, return error
	fnCpuPercent = func(time.Duration, bool) (float64s []float64, e error) {
		ret := make([]float64, 1)
		return ret, errors.New("custom error")
	}
	cpuUsage, err = GetCPUTotalPercentage(100 * time.Millisecond)
	fmt.Println(cpuUsage, err.GetMessage())
	assert(cpuUsage).Equals(float64(0))
	assert(err.GetMessage()).Equals("custom error")
}
