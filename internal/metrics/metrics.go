// Package metrics ...
package metrics

import (
	"fmt"
	"reflect"
	"time"
)

// Metrics ...
type Metrics struct {
	cpu float64
}

// GetMetrics ...
func GetMetrics(interval time.Duration) *Metrics {
	cpu, ok := getCPUPercent(interval)

	if !ok {
		return nil
	}

	return &Metrics{
		cpu: cpu,
	}
}

type cpuTimesStat struct {
	User   float64
	System float64
	Idle   float64
	Nice   float64
}

var checkEqual = func(a, b interface{}) bool {
	switch a.(type) {
	case int:
		return a == b
	default:
		return reflect.DeepEqual(a, b)
	}
}

func (t *cpuTimesStat) getAllBusy() (float64, float64) {
	busy := t.User + t.System + t.Nice
	return busy + t.Idle, busy
}

func calculateCPUPercent(t1 *cpuTimesStat, t2 *cpuTimesStat) (float64, bool) {
	if t1 == nil {
		fmt.Println("TA")
		return 1, false
	}

	if t2 == nil {
		fmt.Println("TB")
		return 1, false
	}

	t1All, t1Busy := t1.getAllBusy()
	t2All, t2Busy := t2.getAllBusy()

	if t2Busy < t1Busy {
		fmt.Println("TC")
		return 1, false
	}

	if t2All <= t1All {
		fmt.Println("TD")
		return 1, false
	}

	return (t2Busy - t1Busy) / (t2All - t1All), true
}

func getCPUPercent(interval time.Duration) (float64, bool) {
	if interval <= 0 {
		return 1, false
	}

	t1 := allCPUTimes()
	time.Sleep(interval)
	t2 := allCPUTimes()
	return calculateCPUPercent(t1, t2)
}
