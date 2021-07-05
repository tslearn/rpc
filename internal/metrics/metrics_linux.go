// +build linux

package metrics

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

var getProcStat = func() string {
	f, err := os.Open("/proc/stat")
	if !checkEqual(err, nil) {
		return ""
	}
	defer func() {
		_ = f.Close()
	}()

	r := bufio.NewReader(f)
	line, _ := r.ReadString('\n')
	return line
}

func parseProcStat(line string) *cpuTimesStat {
	if fields := strings.Fields(line); len(fields) < 5 {
		return nil
	} else if strings.HasPrefix(fields[0], "cpu") == false {
		return nil
	} else if user, err := strconv.ParseFloat(fields[1], 64); err != nil {
		return nil
	} else if nice, err := strconv.ParseFloat(fields[2], 64); err != nil {
		return nil
	} else if system, err := strconv.ParseFloat(fields[3], 64); err != nil {
		return nil
	} else if idle, err := strconv.ParseFloat(fields[4], 64); err != nil {
		return nil
	} else {
		return &cpuTimesStat{
			User:   user,
			Nice:   nice,
			System: system,
			Idle:   idle,
		}
	}
}

func allCPUTimes() *cpuTimesStat {
	return parseProcStat(getProcStat())
}
