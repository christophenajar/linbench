package benchmark

import (
	"runtime"
	"time"

	"linbench/internal/model"
	"linbench/internal/unit"
)

func RunMemory(sizeBytes int64, duration time.Duration) model.MemoryResult {
	sizeBytes = clampSize(sizeBytes, 64*unit.MiB, 2*unit.GiB)
	buf := make([]byte, sizeBytes)
	src := make([]byte, sizeBytes/2)
	dst := make([]byte, sizeBytes/2)
	for i := range src {
		src[i] = byte(i)
	}

	runtime.GC()
	readBPS := timedBandwidth(duration, func() int64 {
		var sum uint64
		for i := 0; i < len(buf); i += 64 {
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

	return model.MemoryResult{
		ReadMBS:   readBPS / unit.MB,
		WriteMBS:  writeBPS / unit.MB,
		CopyMBS:   copyBPS / unit.MB,
		LatencyNS: latencyNS(sizeBytes, duration),
	}
}
