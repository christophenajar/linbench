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
		L1: runCacheLevel(sizes[1], sizes[1]*3/4, duration),
		L2: runCacheLevel(sizes[2], sizes[2]*3/4, duration),
		L3: runCacheLevel(sizes[3], sizes[3]*3/4, duration),
	}
}

func runCacheLevel(cacheSize, benchSize int64, duration time.Duration) model.CacheLevelResult {
	benchSize = clampSize(benchSize, 4*unit.KiB, cacheSize)
	buf := make([]byte, benchSize)
	src := make([]byte, benchSize)
	dst := make([]byte, benchSize)
	for i := range src {
		src[i] = byte(i)
		buf[i] = byte(i)
	}

	runtime.GC()
	readBPS := timedBandwidth(duration, func() int64 {
		var sum uint64
		for i := 0; i < len(buf); i += 8 {
			sum += uint64(buf[i])
		}
		Sink = sum
		return int64(len(buf))
	})

	runtime.GC()
	writeBPS := timedBandwidth(duration, func() int64 {
		for i := range buf {
			buf[i] = byte(i)
		}
		Sink = uint64(buf[len(buf)-1])
		return int64(len(buf))
	})

	runtime.GC()
	copyBPS := timedBandwidth(duration, func() int64 {
		n := copy(dst, src)
		Sink = uint64(dst[n-1])
		return int64(n)
	})

	return model.CacheLevelResult{
		SizeBytes: cacheSize,
		ReadGBS:   readBPS / unit.GB,
		WriteGBS:  writeBPS / unit.GB,
		CopyGBS:   copyBPS / unit.GB,
		LatencyNS: latencyNS(benchSize, duration),
	}
}
