package benchmark

import (
	"crypto/rand"
	"fmt"
	"io"
	mrand "math/rand"
	"os"
	"path/filepath"
	"time"

	"linbench/internal/model"
	"linbench/internal/unit"
)

type DiskOptions struct {
	Path         string
	SizeBytes    int64
	Duration     time.Duration
	KeepTestFile bool
	NoSync       bool
}

func RunDisk(options DiskOptions) (model.DiskResult, error) {
	if options.Path == "" {
		options.Path = os.TempDir()
	}
	if options.SizeBytes <= 0 {
		options.SizeBytes = 256 * unit.MiB
	}
	if options.Duration <= 0 {
		options.Duration = time.Second
	}

	file, err := os.CreateTemp(options.Path, "linbench-*.tmp")
	if err != nil {
		return model.DiskResult{Path: options.Path}, err
	}
	path := file.Name()
	if !options.KeepTestFile {
		defer os.Remove(path)
	}
	defer file.Close()

	result := model.DiskResult{Path: filepath.Dir(path)}
	writeMBS, err := sequentialWrite(file, options.SizeBytes, options.NoSync)
	if err != nil {
		return result, err
	}
	result.SequentialWriteMBS = writeMBS
	fadviseDontNeed(file)

	readMBS, err := sequentialRead(file)
	if err != nil {
		return result, err
	}
	result.SequentialReadMBS = readMBS
	fadviseDontNeed(file)

	readIOPS, latencyMS, err := randomRead(file, options.SizeBytes, options.Duration)
	if err != nil {
		return result, err
	}
	result.RandomReadIOPS = readIOPS
	result.LatencyMS = latencyMS

	writeIOPS, err := randomWrite(file, options.SizeBytes, options.Duration, options.NoSync)
	if err != nil {
		return result, err
	}
	result.RandomWriteIOPS = writeIOPS
	return result, nil
}

func sequentialWrite(file *os.File, sizeBytes int64, noSync bool) (float64, error) {
	buf := make([]byte, unit.MiB)
	if _, err := rand.Read(buf); err != nil {
		return 0, err
	}
	if err := file.Truncate(0); err != nil {
		return 0, err
	}
	if _, err := file.Seek(0, 0); err != nil {
		return 0, err
	}

	start := time.Now()
	var written int64
	for written < sizeBytes {
		chunk := int64(len(buf))
		if remain := sizeBytes - written; remain < chunk {
			chunk = remain
		}
		n, err := file.Write(buf[:chunk])
		if err != nil {
			return 0, err
		}
		written += int64(n)
	}
	if !noSync {
		if err := file.Sync(); err != nil {
			return 0, err
		}
	}
	elapsed := time.Since(start).Seconds()
	if elapsed <= 0 {
		return 0, nil
	}
	return float64(written) / elapsed / unit.MB, nil
}

func sequentialRead(file *os.File) (float64, error) {
	if _, err := file.Seek(0, 0); err != nil {
		return 0, err
	}
	buf := make([]byte, unit.MiB)
	start := time.Now()
	var total int64
	for {
		n, err := file.Read(buf)
		total += int64(n)
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
	}
	elapsed := time.Since(start).Seconds()
	if elapsed <= 0 {
		return 0, nil
	}
	return float64(total) / elapsed / unit.MB, nil
}

func randomRead(file *os.File, sizeBytes int64, duration time.Duration) (float64, float64, error) {
	const blockSize = 4096
	if sizeBytes < blockSize {
		return 0, 0, fmt.Errorf("disk test file too small")
	}
	buf := make([]byte, blockSize)
	rng := mrand.New(mrand.NewSource(1))
	deadline := time.Now().Add(duration)
	var ops int64
	start := time.Now()
	for time.Now().Before(deadline) {
		offset := randomOffset(rng, sizeBytes, blockSize)
		if _, err := file.ReadAt(buf, offset); err != nil {
			return 0, 0, err
		}
		ops++
	}
	elapsed := time.Since(start).Seconds()
	if elapsed <= 0 || ops == 0 {
		return 0, 0, nil
	}
	iops := float64(ops) / elapsed
	latencyMS := elapsed * 1000 / float64(ops)
	return iops, latencyMS, nil
}

func randomWrite(file *os.File, sizeBytes int64, duration time.Duration, noSync bool) (float64, error) {
	const blockSize = 4096
	if sizeBytes < blockSize {
		return 0, fmt.Errorf("disk test file too small")
	}
	buf := make([]byte, blockSize)
	rng := mrand.New(mrand.NewSource(2))
	deadline := time.Now().Add(duration)
	var ops int64
	start := time.Now()
	for time.Now().Before(deadline) {
		offset := randomOffset(rng, sizeBytes, blockSize)
		if _, err := file.WriteAt(buf, offset); err != nil {
			return 0, err
		}
		ops++
	}
	if !noSync {
		if err := file.Sync(); err != nil {
			return 0, err
		}
	}
	elapsed := time.Since(start).Seconds()
	if elapsed <= 0 {
		return 0, nil
	}
	return float64(ops) / elapsed, nil
}

func randomOffset(rng *mrand.Rand, sizeBytes int64, blockSize int64) int64 {
	blocks := sizeBytes / blockSize
	if blocks <= 1 {
		return 0
	}
	return rng.Int63n(blocks) * blockSize
}
