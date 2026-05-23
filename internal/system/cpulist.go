package system

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func ParseCPUList(value string) map[int]bool {
	cpus := map[int]bool{}
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			start, errStart := strconv.Atoi(strings.TrimSpace(bounds[0]))
			end, errEnd := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if errStart != nil || errEnd != nil || end < start {
				continue
			}
			for cpu := start; cpu <= end; cpu++ {
				cpus[cpu] = true
			}
			continue
		}
		cpu, err := strconv.Atoi(part)
		if err == nil {
			cpus[cpu] = true
		}
	}
	return cpus
}

func CPUListContainsRemote(allowed string, local string) bool {
	allowedSet := ParseCPUList(allowed)
	localSet := ParseCPUList(local)
	if len(allowedSet) == 0 || len(localSet) == 0 {
		return false
	}
	for cpu := range allowedSet {
		if !localSet[cpu] {
			return true
		}
	}
	return false
}

func CPUListHasIntersection(a string, b string) bool {
	aSet := ParseCPUList(a)
	bSet := ParseCPUList(b)
	for cpu := range aSet {
		if bSet[cpu] {
			return true
		}
	}
	return false
}

func FormatCPUSet(cpus map[int]bool) string {
	values := make([]int, 0, len(cpus))
	for cpu := range cpus {
		values = append(values, cpu)
	}
	sort.Ints(values)
	parts := make([]string, 0, len(values))
	for i := 0; i < len(values); {
		start := values[i]
		end := start
		for i+1 < len(values) && values[i+1] == end+1 {
			i++
			end = values[i]
		}
		if start == end {
			parts = append(parts, strconv.Itoa(start))
		} else {
			parts = append(parts, fmt.Sprintf("%d-%d", start, end))
		}
		i++
	}
	return strings.Join(parts, ",")
}
