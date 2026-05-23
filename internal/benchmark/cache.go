package benchmark

import (
	"runtime"
	"time"

	"linbench/internal/model"
	"linbench/internal/system"
	"linbench/internal/unit"
)

func RunCache(duration time.Duration) model.CacheResult {
	caches, _ := system.ReadCacheInfo()
	sizes := system.CacheSizesByLevel(caches)
	if sizes[1] == 0 {
		sizes[1] = 32 * unit.KiB
	}
	if sizes[2] == 0 {
		sizes[2] = 256 * unit.KiB
	}
	if sizes[3] == 0 {
		sizes[3] = 8 * unit.MiB
	}

	return model.CacheResult{
		L1: runCacheLevel(sizes[1], cacheBenchSize(0, sizes[1]), duration),
		L2: runCacheLevel(sizes[2], cacheBenchSize(sizes[1], sizes[2]), duration),
		L3: runCacheLevel(sizes[3], cacheBenchSize(sizes[2], sizes[3]), duration),
	}
}

func cacheBenchSize(lowerLevelSize, cacheSize int64) int64 {
	target := cacheSize * 3 / 4
	if lowerLevelSize > 0 {
		min := lowerLevelSize * 2
		if target < min && min < cacheSize {
			target = min
		}
	}
	return target
}

func runCacheLevel(cacheSize, benchSize int64, duration time.Duration) model.CacheLevelResult {
	benchSize = clampSize(benchSize, 4*unit.KiB, cacheSize)
	words := int(benchSize / 8)
	if words < 1024 {
		words = 1024
	}
	buf := make([]uint64, words)
	src := make([]uint64, words)
	dst := make([]uint64, words)
	fillUint64(buf)
	fillUint64(src)

	runtime.GC()
	readBPS := readBandwidth(buf, duration)

	runtime.GC()
	writeBPS := writeBandwidth(buf, duration)

	runtime.GC()
	copyBPS := copyBandwidth(dst, src, duration)

	return model.CacheLevelResult{
		SizeBytes: cacheSize,
		ReadGBS:   readBPS / unit.GB,
		WriteGBS:  writeBPS / unit.GB,
		CopyGBS:   copyBPS / unit.GB,
		LatencyNS: latencyNS(benchSize, duration),
	}
}
