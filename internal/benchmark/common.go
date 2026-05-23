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
	const minEntries = 1024
	count := int(sizeBytes / 64)
	if count < minEntries {
		count = minEntries
	}
	indices := rand.New(rand.NewSource(1)).Perm(count)
	next := make([]int, count)
	for i := 0; i < count-1; i++ {
		next[indices[i]] = indices[i+1]
	}
	next[indices[count-1]] = indices[0]

	if duration <= 0 {
		duration = 200 * time.Millisecond
	}
	runtime.GC()
	deadline := time.Now().Add(duration)
	idx := indices[0]
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

func clampSize(size, min, max int64) int64 {
	if size < min {
		return min
	}
	if max > 0 && size > max {
		return max
	}
	return size
}
