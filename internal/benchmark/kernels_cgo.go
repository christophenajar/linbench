//go:build cgo

package benchmark

/*
#include <stdint.h>
#include <stddef.h>
#include <string.h>
#include <time.h>

static volatile uint64_t linbench_sink;

static uint64_t linbench_now_ns(void) {
	struct timespec ts;
	clock_gettime(CLOCK_MONOTONIC, &ts);
	return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}

static double linbench_elapsed_bps(uint64_t bytes, uint64_t start_ns) {
	uint64_t elapsed_ns = linbench_now_ns() - start_ns;
	if (elapsed_ns == 0) {
		return 0.0;
	}
	return ((double)bytes * 1000000000.0) / (double)elapsed_ns;
}

static double linbench_read_u64(uint64_t *buf, size_t n, int64_t duration_ns) {
	if (n == 0 || duration_ns <= 0) {
		return 0.0;
	}
	uint64_t start = linbench_now_ns();
	uint64_t deadline = start + (uint64_t)duration_ns;
	uint64_t bytes = 0;
	uint64_t sum = 0;
	do {
		for (size_t i = 0; i < n; i++) {
			sum += buf[i];
		}
		bytes += (uint64_t)n * 8ULL;
	} while (linbench_now_ns() < deadline);
	linbench_sink = sum;
	return linbench_elapsed_bps(bytes, start);
}

static double linbench_write_u64(uint64_t *buf, size_t n, int64_t duration_ns) {
	if (n == 0 || duration_ns <= 0) {
		return 0.0;
	}
	uint64_t start = linbench_now_ns();
	uint64_t deadline = start + (uint64_t)duration_ns;
	uint64_t bytes = 0;
	uint64_t iter = 0;
	do {
		for (size_t i = 0; i < n; i++) {
			buf[i] = ((uint64_t)i * 0x9e3779b97f4a7c15ULL) + iter;
		}
		bytes += (uint64_t)n * 8ULL;
		iter++;
	} while (linbench_now_ns() < deadline);
	linbench_sink = buf[n - 1];
	return linbench_elapsed_bps(bytes, start);
}

static double linbench_copy_u64(uint64_t *dst, uint64_t *src, size_t n, int64_t duration_ns) {
	if (n == 0 || duration_ns <= 0) {
		return 0.0;
	}
	uint64_t start = linbench_now_ns();
	uint64_t deadline = start + (uint64_t)duration_ns;
	uint64_t bytes = 0;
	size_t byte_count = n * sizeof(uint64_t);
	do {
		memcpy(dst, src, byte_count);
		bytes += (uint64_t)byte_count;
	} while (linbench_now_ns() < deadline);
	linbench_sink = dst[n - 1];
	return linbench_elapsed_bps(bytes, start);
}
*/
import "C"

import (
	"time"
	"unsafe"
)

func readBandwidth(values []uint64, duration time.Duration) float64 {
	if len(values) == 0 {
		return 0
	}
	return float64(C.linbench_read_u64(
		(*C.uint64_t)(unsafe.Pointer(&values[0])),
		C.size_t(len(values)),
		C.int64_t(duration.Nanoseconds()),
	))
}

func writeBandwidth(values []uint64, duration time.Duration) float64 {
	if len(values) == 0 {
		return 0
	}
	return float64(C.linbench_write_u64(
		(*C.uint64_t)(unsafe.Pointer(&values[0])),
		C.size_t(len(values)),
		C.int64_t(duration.Nanoseconds()),
	))
}

func copyBandwidth(dst, src []uint64, duration time.Duration) float64 {
	if len(dst) == 0 || len(src) == 0 {
		return 0
	}
	n := len(src)
	if len(dst) < n {
		n = len(dst)
	}
	return float64(C.linbench_copy_u64(
		(*C.uint64_t)(unsafe.Pointer(&dst[0])),
		(*C.uint64_t)(unsafe.Pointer(&src[0])),
		C.size_t(n),
		C.int64_t(duration.Nanoseconds()),
	))
}
