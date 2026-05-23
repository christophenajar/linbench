//go:build !cgo

package benchmark

import "time"

func EngineName() string {
	return "go"
}

func readBandwidth(values []uint64, duration time.Duration) float64 {
	return timedBandwidth(duration, func() int64 {
		var sum uint64
		for _, value := range values {
			sum += value
		}
		Sink = sum
		return int64(len(values) * 8)
	})
}

func writeBandwidth(values []uint64, duration time.Duration) float64 {
	return timedBandwidth(duration, func() int64 {
		for i := range values {
			values[i] = uint64(i) * 0x9e3779b97f4a7c15
		}
		Sink = values[len(values)-1]
		return int64(len(values) * 8)
	})
}

func copyBandwidth(dst, src []uint64, duration time.Duration) float64 {
	return timedBandwidth(duration, func() int64 {
		n := copy(dst, src)
		Sink = dst[n-1]
		return int64(n * 8)
	})
}
