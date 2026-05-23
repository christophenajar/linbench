package benchmark

import (
	"compress/gzip"
	"crypto/sha256"
	"io"
	"math"
	"runtime"
	"sync"
	"time"

	"linbench/internal/model"
	"linbench/internal/unit"
)

func RunCPU(threads int, duration time.Duration) model.CPUResult {
	if threads <= 0 {
		threads = runtime.NumCPU()
	}
	if duration <= 0 {
		duration = time.Second
	}
	oldProcs := runtime.GOMAXPROCS(threads)
	defer runtime.GOMAXPROCS(oldProcs)

	return model.CPUResult{
		IntegerScore:   runParallel(threads, duration, integerWorker),
		FloatScore:     runParallel(threads, duration, floatWorker),
		SHA256MBS:      runParallelBytes(threads, duration, sha256Worker) / unit.MB,
		CompressionMBS: runParallelBytes(threads, duration, compressionWorker) / unit.MB,
	}
}

func runParallel(threads int, duration time.Duration, worker func(time.Time) float64) float64 {
	deadline := time.Now().Add(duration)
	var wg sync.WaitGroup
	results := make(chan float64, threads)
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- worker(deadline)
		}()
	}
	wg.Wait()
	close(results)
	var total float64
	for value := range results {
		total += value
	}
	return total / duration.Seconds()
}

func runParallelBytes(threads int, duration time.Duration, worker func(time.Time) int64) float64 {
	deadline := time.Now().Add(duration)
	var wg sync.WaitGroup
	results := make(chan int64, threads)
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- worker(deadline)
		}()
	}
	wg.Wait()
	close(results)
	var total int64
	for value := range results {
		total += value
	}
	return float64(total) / duration.Seconds()
}

func integerWorker(deadline time.Time) float64 {
	var x uint64 = 0x9e3779b97f4a7c15
	var ops float64
	for time.Now().Before(deadline) {
		for i := 0; i < 10000; i++ {
			x += uint64(i)
			x ^= x << 13
			x *= 0xbf58476d1ce4e5b9
			x ^= x >> 17
			ops += 4
		}
	}
	Sink = x
	return ops
}

func floatWorker(deadline time.Time) float64 {
	x := 1.000001
	y := 1.0000003
	var ops float64
	for time.Now().Before(deadline) {
		for i := 0; i < 10000; i++ {
			x += y
			y *= 1.0000001
			x /= 1.00000001
			y = math.Sqrt(y + 1.0)
			ops += 4
		}
	}
	Sink = math.Float64bits(x + y)
	return ops
}

func sha256Worker(deadline time.Time) int64 {
	buf := make([]byte, unit.MiB)
	for i := range buf {
		buf[i] = byte(i)
	}
	var total int64
	for time.Now().Before(deadline) {
		sum := sha256.Sum256(buf)
		Sink = uint64(sum[0])
		total += int64(len(buf))
	}
	return total
}

func compressionWorker(deadline time.Time) int64 {
	buf := make([]byte, unit.MiB)
	for i := range buf {
		buf[i] = byte(i)
	}
	var total int64
	for time.Now().Before(deadline) {
		zw := gzip.NewWriter(io.Discard)
		if _, err := zw.Write(buf); err != nil {
			break
		}
		if err := zw.Close(); err != nil {
			break
		}
		total += int64(len(buf))
	}
	return total
}
