package system

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

type MemInfo struct {
	TotalBytes     int64
	AvailableBytes int64
}

func ReadMemInfo() MemInfo {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return MemInfo{}
	}
	defer f.Close()
	return ParseMemInfo(f)
}

func ParseMemInfo(r io.Reader) MemInfo {
	var info MemInfo
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 2 {
			continue
		}
		value, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}
		bytes := value * 1024
		switch strings.TrimSuffix(parts[0], ":") {
		case "MemTotal":
			info.TotalBytes = bytes
		case "MemAvailable":
			info.AvailableBytes = bytes
		}
	}
	return info
}
