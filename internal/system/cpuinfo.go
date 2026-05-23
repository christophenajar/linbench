package system

import (
	"bufio"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"

	"linbench/internal/model"
)

func ReadSystemInfo() model.SystemInfo {
	host, _ := os.Hostname()
	info := model.SystemInfo{
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		Hostname:   host,
		Kernel:     kernelRelease(),
		CPUThreads: runtime.NumCPU(),
	}

	if f, err := os.Open("/proc/cpuinfo"); err == nil {
		defer f.Close()
		modelName, cores, threads := ParseCPUInfo(f)
		info.CPUModel = modelName
		if cores > 0 {
			info.CPUCores = cores
		}
		if threads > 0 {
			info.CPUThreads = threads
		}
	}
	if info.CPUCores == 0 {
		info.CPUCores = info.CPUThreads
	}
	if info.CPUModel == "" {
		info.CPUModel = "unknown"
	}
	return info
}

func ParseCPUInfo(r io.Reader) (modelName string, cores int, threads int) {
	scanner := bufio.NewScanner(r)
	physicalIDs := map[string]bool{}
	coreIDs := map[string]bool{}
	currentPhysical := "0"

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "processor":
			threads++
		case "model name", "Hardware":
			if modelName == "" {
				modelName = value
			}
		case "cpu cores":
			if cores == 0 {
				cores, _ = strconv.Atoi(value)
			}
		case "physical id":
			currentPhysical = value
			physicalIDs[value] = true
		case "core id":
			coreIDs[currentPhysical+":"+value] = true
		}
	}

	if len(coreIDs) > 0 {
		cores = len(coreIDs)
	} else if cores > 0 && len(physicalIDs) > 1 {
		cores *= len(physicalIDs)
	}
	return modelName, cores, threads
}

func kernelRelease() string {
	value, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return "unknown"
	}
	release := strings.TrimSpace(string(value))
	if release == "" {
		return "unknown"
	}
	return release
}
