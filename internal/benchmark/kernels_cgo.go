//go:build cgo

package benchmark

/*
#cgo CFLAGS: -O3
#include <stdint.h>
#include <stddef.h>
#include <string.h>
#include <time.h>
#if defined(__x86_64__) || defined(__i386__)
#include <immintrin.h>
#endif

static volatile uint64_t linbench_sink;
static const uint64_t linbench_check_bytes = 64ULL * 1024ULL * 1024ULL;

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

static size_t linbench_batch_passes(size_t byte_count) {
	if (byte_count == 0) {
		return 1;
	}
	size_t passes = (size_t)(linbench_check_bytes / (uint64_t)byte_count);
	if (passes < 1) {
		return 1;
	}
	if (passes > 65536) {
		return 65536;
	}
	return passes;
}

static double linbench_read_u64(uint64_t *buf, size_t n, int64_t duration_ns) {
	if (n == 0 || duration_ns <= 0) {
		return 0.0;
	}
	uint64_t start = linbench_now_ns();
	uint64_t deadline = start + (uint64_t)duration_ns;
	uint64_t bytes = 0;
	uint64_t s0 = 0;
	uint64_t s1 = 0;
	uint64_t s2 = 0;
	uint64_t s3 = 0;
	uint64_t s4 = 0;
	uint64_t s5 = 0;
	uint64_t s6 = 0;
	uint64_t s7 = 0;
	size_t batch_passes = linbench_batch_passes(n * sizeof(uint64_t));
	do {
		for (size_t pass = 0; pass < batch_passes; pass++) {
			size_t i = 0;
			for (; i + 7 < n; i += 8) {
				s0 += buf[i + 0];
				s1 += buf[i + 1];
				s2 += buf[i + 2];
				s3 += buf[i + 3];
				s4 += buf[i + 4];
				s5 += buf[i + 5];
				s6 += buf[i + 6];
				s7 += buf[i + 7];
			}
			for (; i < n; i++) {
				s0 += buf[i];
			}
		}
		bytes += (uint64_t)n * 8ULL * (uint64_t)batch_passes;
	} while (linbench_now_ns() < deadline);
	linbench_sink = s0 + s1 + s2 + s3 + s4 + s5 + s6 + s7;
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
	size_t batch_passes = linbench_batch_passes(n * sizeof(uint64_t));
	do {
		for (size_t pass = 0; pass < batch_passes; pass++) {
			size_t i = 0;
#if defined(__AVX2__)
			__m256i v0 = _mm256_set_epi64x(
				(int64_t)(0x1111111111111111ULL + iter),
				(int64_t)(0x2222222222222222ULL + iter),
				(int64_t)(0x3333333333333333ULL + iter),
				(int64_t)(0x4444444444444444ULL + iter));
			__m256i v1 = _mm256_set_epi64x(
				(int64_t)(0x5555555555555555ULL + iter),
				(int64_t)(0x6666666666666666ULL + iter),
				(int64_t)(0x7777777777777777ULL + iter),
				(int64_t)(0x8888888888888888ULL + iter));
			__m256i v2 = _mm256_set_epi64x(
				(int64_t)(0x9999999999999999ULL + iter),
				(int64_t)(0xaaaaaaaaaaaaaaaaULL + iter),
				(int64_t)(0xbbbbbbbbbbbbbbbbULL + iter),
				(int64_t)(0xccccccccccccccccULL + iter));
			__m256i v3 = _mm256_set_epi64x(
				(int64_t)(0xddddddddddddddddULL + iter),
				(int64_t)(0xeeeeeeeeeeeeeeeeULL + iter),
				(int64_t)(0xf0f0f0f0f0f0f0f0ULL + iter),
				(int64_t)(0x0f0f0f0f0f0f0f0fULL + iter));
			for (; i + 15 < n; i += 16) {
				_mm256_storeu_si256((__m256i *)(buf + i + 0), v0);
				_mm256_storeu_si256((__m256i *)(buf + i + 4), v1);
				_mm256_storeu_si256((__m256i *)(buf + i + 8), v2);
				_mm256_storeu_si256((__m256i *)(buf + i + 12), v3);
			}
#endif
			for (; i + 7 < n; i += 8) {
				buf[i + 0] = 0x1111111111111111ULL + iter;
				buf[i + 1] = 0x2222222222222222ULL + iter;
				buf[i + 2] = 0x3333333333333333ULL + iter;
				buf[i + 3] = 0x4444444444444444ULL + iter;
				buf[i + 4] = 0x5555555555555555ULL + iter;
				buf[i + 5] = 0x6666666666666666ULL + iter;
				buf[i + 6] = 0x7777777777777777ULL + iter;
				buf[i + 7] = 0x8888888888888888ULL + iter;
			}
			for (; i < n; i++) {
				buf[i] = 0x9e3779b97f4a7c15ULL + iter;
			}
			iter++;
		}
		bytes += (uint64_t)n * 8ULL * (uint64_t)batch_passes;
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
	size_t batch_passes = linbench_batch_passes(byte_count);
	do {
		for (size_t pass = 0; pass < batch_passes; pass++) {
			memcpy(dst, src, byte_count);
		}
		bytes += (uint64_t)byte_count * (uint64_t)batch_passes;
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

func EngineName() string {
	return "cgo"
}

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
