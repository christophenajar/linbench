package benchmark

import (
	"runtime"
	"time"

	"linbench/internal/model"
	"linbench/internal/unit"
)

func RunMemory(sizeBytes int64, duration time.Duration) model.MemoryResult {
	sizeBytes = clampSize(sizeBytes, 64*unit.MiB, 2*unit.GiB)
	words := int(sizeBytes / 8)
	copyWords := words / 2
	buf := make([]uint64, words)
	src := make([]uint64, copyWords)
	dst := make([]uint64, copyWords)
	fillUint64(buf)
	fillUint64(src)

	runtime.GC()
	readBPS := readBandwidth(buf, duration)

	runtime.GC()
	writeBPS := writeBandwidth(buf, duration)

	runtime.GC()
	copyBPS := copyBandwidth(dst, src, duration)

	buf = nil
	src = nil
	dst = nil
	runtime.GC()
	latency := latencyNS(sizeBytes, duration)

	return model.MemoryResult{
		ReadMBS:   readBPS / unit.MB,
		WriteMBS:  writeBPS / unit.MB,
		CopyMBS:   copyBPS / unit.MB,
		LatencyNS: latency,
	}
}
