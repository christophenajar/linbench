package system

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"linbench/internal/model"
)

func ReadNUMAInfo() model.NUMAInfo {
	dirs, err := filepath.Glob("/sys/devices/system/node/node[0-9]*")
	if err != nil || len(dirs) == 0 {
		return model.NUMAInfo{}
	}

	nodes := make([]model.NUMANodeInfo, 0, len(dirs))
	for _, dir := range dirs {
		id, ok := parseNodeID(filepath.Base(dir))
		if !ok {
			continue
		}
		node := model.NUMANodeInfo{
			ID:       id,
			CPUs:     readTrimmed(filepath.Join(dir, "cpulist")),
			Distance: parseDistanceLine(readTrimmed(filepath.Join(dir, "distance"))),
		}
		if f, err := os.Open(filepath.Join(dir, "meminfo")); err == nil {
			node.MemTotalBytes, node.MemFreeBytes = parseNUMANodeMemInfo(f)
			f.Close()
		}
		nodes = append(nodes, node)
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})
	return model.NUMAInfo{
		Available: len(nodes) > 0,
		NodeCount: len(nodes),
		Nodes:     nodes,
	}
}

func parseNodeID(name string) (int, bool) {
	if !strings.HasPrefix(name, "node") {
		return 0, false
	}
	id, err := strconv.Atoi(strings.TrimPrefix(name, "node"))
	return id, err == nil
}

func parseDistanceLine(line string) []int {
	fields := strings.Fields(line)
	values := make([]int, 0, len(fields))
	for _, field := range fields {
		value, err := strconv.Atoi(field)
		if err == nil {
			values = append(values, value)
		}
	}
	return values
}

func parseNUMANodeMemInfo(r io.Reader) (totalBytes int64, freeBytes int64) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		for i := 0; i+1 < len(fields); i++ {
			key := strings.TrimSuffix(fields[i], ":")
			if key != "MemTotal" && key != "MemFree" {
				continue
			}
			value, err := strconv.ParseInt(fields[i+1], 10, 64)
			if err != nil {
				continue
			}
			switch key {
			case "MemTotal":
				totalBytes = value * 1024
			case "MemFree":
				freeBytes = value * 1024
			}
		}
	}
	return totalBytes, freeBytes
}

func readTrimmed(path string) string {
	value, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(value))
}
