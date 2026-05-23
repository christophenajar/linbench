package system

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"linbench/internal/unit"
)

type CacheInfo struct {
	Level     int
	Type      string
	SizeBytes int64
	LineSize  int64
	Ways      int
}

func ReadCacheInfo() ([]CacheInfo, error) {
	dirs, err := filepath.Glob("/sys/devices/system/cpu/cpu0/cache/index*")
	if err != nil {
		return nil, err
	}
	var caches []CacheInfo
	for _, dir := range dirs {
		level := readInt(filepath.Join(dir, "level"))
		typ := readString(filepath.Join(dir, "type"))
		size := readCacheSize(filepath.Join(dir, "size"))
		line := readInt64(filepath.Join(dir, "coherency_line_size"))
		ways := readInt(filepath.Join(dir, "ways_of_associativity"))
		if level > 0 && size > 0 {
			caches = append(caches, CacheInfo{
				Level:     level,
				Type:      typ,
				SizeBytes: size,
				LineSize:  line,
				Ways:      ways,
			})
		}
	}
	return caches, nil
}

func CacheSizesByLevel(caches []CacheInfo) map[int]int64 {
	sizes := map[int]int64{}
	for _, cache := range caches {
		if cache.Level == 1 && cache.Type != "Data" && cache.Type != "Unified" {
			continue
		}
		if cache.SizeBytes > sizes[cache.Level] {
			sizes[cache.Level] = cache.SizeBytes
		}
	}
	return sizes
}

func ParseCacheSize(s string) (int64, error) {
	return unit.ParseSize(strings.TrimSpace(s))
}

func readCacheSize(path string) int64 {
	value, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	size, _ := ParseCacheSize(string(value))
	return size
}

func readString(path string) string {
	value, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(value))
}

func readInt(path string) int {
	value, _ := strconv.Atoi(readString(path))
	return value
}

func readInt64(path string) int64 {
	value, _ := strconv.ParseInt(readString(path), 10, 64)
	return value
}
