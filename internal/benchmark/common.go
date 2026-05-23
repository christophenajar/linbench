package benchmark

import (
	"math/rand"
	"runtime"
	"time"
)

var Sink uint64

func timedBandwidth(duration time.Duration, fn func() int64) float64 {
	if duration <= 0 {
		duration = 200 * time.Millisecond
	}
	deadline := time.Now().Add(duration)
	var bytes int64
	start := time.Now()
	for {
		bytes += fn()
		if time.Now().After(deadline) {
			break
		}
	}
	elapsed := time.Since(start).Seconds()
	if elapsed <= 0 {
		return 0
	}
	return float64(bytes) / elapsed
}

func latencyNS(sizeBytes int64, duration time.Duration) float64 {
	const (
		minEntries            = 1024
		maxLatencyWorkingSet  = 128 * 1024 * 1024
		pointerChaseWordBytes = 4
	)
	if sizeBytes > maxLatencyWorkingSet {
		sizeBytes = maxLatencyWorkingSet
	}
	count := int(sizeBytes / pointerChaseWordBytes)
	if count < minEntries {
		count = minEntries
	}
	rng := rand.New(rand.NewSource(1))
	indices := make([]uint32, count)
	for i := range indices {
		indices[i] = uint32(i)
	}
	rng.Shuffle(count, func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	next := make([]uint32, count)
	for i := 0; i < count-1; i++ {
		next[indices[i]] = indices[i+1]
	}
	next[indices[count-1]] = indices[0]

	if duration <= 0 {
		duration = 200 * time.Millisecond
	}
	idx := indices[0]
	indices = nil
	runtime.GC()
	deadline := time.Now().Add(duration)
	accesses := 0
	start := time.Now()
	for {
		for i := 0; i < count; i++ {
			idx = next[idx]
		}
		accesses += count
		if time.Now().After(deadline) {
			break
		}
	}
	Sink = uint64(idx)
	if accesses == 0 {
		return 0
	}
	return float64(time.Since(start).Nanoseconds()) / float64(accesses)
}

func fillUint64(values []uint64) {
	var x uint64 = 0x9e3779b97f4a7c15
	for i := range values {
		x ^= x >> 12
		x ^= x << 25
		x ^= x >> 27
		values[i] = x * 0x2545f4914f6cdd1d
	}
}

func clampSize(size, min, max int64) int64 {
	if size < min {
		return min
	}
	if max > 0 && size > max {
		return max
	}
	return size
}
