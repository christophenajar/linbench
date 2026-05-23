package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"linbench/internal/benchmark"
	"linbench/internal/model"
	"linbench/internal/output"
	"linbench/internal/system"
	"linbench/internal/unit"
)

const version = "0.1.2"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	var (
		allFlag      bool
		ramFlag      bool
		cacheFlag    bool
		cpuFlag      bool
		diskFlag     bool
		jsonFlag     bool
		outputFile   string
		durationText string
		threadsText  string
		memoryText   string
		diskPath     string
		diskText     string
		keepFile     bool
		noSync       bool
		verbose      bool
		versionFlag  bool
	)

	fs := flag.NewFlagSet("linbench", flag.ContinueOnError)
	fs.BoolVar(&allFlag, "all", false, "run all benchmarks")
	fs.BoolVar(&ramFlag, "ram", false, "run memory benchmark")
	fs.BoolVar(&cacheFlag, "cache", false, "run CPU cache benchmark")
	fs.BoolVar(&cpuFlag, "cpu", false, "run CPU benchmark")
	fs.BoolVar(&diskFlag, "disk", false, "run disk benchmark")
	fs.BoolVar(&jsonFlag, "json", false, "write JSON output")
	fs.StringVar(&outputFile, "output", "", "write output to file")
	fs.StringVar(&durationText, "duration", "1s", "approximate duration of each test")
	fs.StringVar(&threadsText, "threads", "auto", "CPU benchmark thread count or auto")
	fs.StringVar(&memoryText, "memory-size", "", "memory buffer size, e.g. 512M or 1G")
	fs.StringVar(&diskPath, "disk-path", defaultDiskPath(), "directory for disk test file")
	fs.StringVar(&diskText, "disk-size", "256M", "disk test file size")
	fs.BoolVar(&keepFile, "keep-test-file", false, "keep disk benchmark temporary file")
	fs.BoolVar(&noSync, "no-sync", false, "skip fsync during disk write tests")
	fs.BoolVar(&verbose, "verbose", false, "print technical details to stderr")
	fs.BoolVar(&versionFlag, "version", false, "print version")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if versionFlag {
		fmt.Println("linbench", version)
		return 0
	}

	sections := model.Sections{
		Memory: ramFlag,
		Cache:  cacheFlag,
		CPU:    cpuFlag,
		Disk:   diskFlag,
	}
	if allFlag || (!ramFlag && !cacheFlag && !cpuFlag && !diskFlag) {
		sections = model.Sections{Memory: true, Cache: true, CPU: true, Disk: true}
	}

	duration, err := time.ParseDuration(durationText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid --duration: %v\n", err)
		return 2
	}
	threads, err := parseThreads(threadsText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid --threads: %v\n", err)
		return 2
	}
	memorySize, err := chooseMemorySize(memoryText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid --memory-size: %v\n", err)
		return 2
	}
	diskSize, err := unit.ParseSize(diskText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid --disk-size: %v\n", err)
		return 2
	}

	result := model.BenchmarkResult{System: system.ReadSystemInfo()}
	if verbose {
		fmt.Fprintf(os.Stderr, "duration=%s threads=%d memory=%d disk=%d\n", duration, threads, memorySize, diskSize)
	}

	if sections.Memory {
		result.Memory = benchmark.RunMemory(memorySize, duration)
	}
	if sections.Cache {
		result.Cache = benchmark.RunCache(duration)
	}
	if sections.CPU {
		result.CPU = benchmark.RunCPU(threads, duration)
	}
	if sections.Disk {
		result.Warnings = append(result.Warnings, diskWarnings(diskPath, diskSize, noSync)...)
		disk, err := benchmark.RunDisk(benchmark.DiskOptions{
			Path:         diskPath,
			SizeBytes:    diskSize,
			Duration:     duration,
			KeepTestFile: keepFile,
			NoSync:       noSync,
		})
		result.Disk = disk
		if err != nil {
			result.Errors = append(result.Errors, "disk: "+err.Error())
		}
	}

	var data []byte
	if jsonFlag {
		data, err = output.FormatJSON(result)
	} else {
		data = []byte(output.FormatText(result, sections))
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "format output: %v\n", err)
		return 1
	}
	data = append(data, '\n')

	if outputFile != "" {
		if err := os.WriteFile(outputFile, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "write output: %v\n", err)
			return 1
		}
		return 0
	}
	if _, err := os.Stdout.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "write stdout: %v\n", err)
		return 1
	}
	return 0
}

func parseThreads(value string) (int, error) {
	if value == "" || value == "auto" {
		return runtime.NumCPU(), nil
	}
	threads, err := strconv.Atoi(value)
	if err != nil || threads <= 0 {
		return 0, fmt.Errorf("expected auto or positive integer")
	}
	return threads, nil
}

func chooseMemorySize(value string) (int64, error) {
	if value != "" {
		return unit.ParseSize(value)
	}
	mem := system.ReadMemInfo()
	if mem.AvailableBytes > 0 {
		size := mem.AvailableBytes / 4
		if size > unit.GiB {
			size = unit.GiB
		}
		if size < 256*unit.MiB {
			size = 256 * unit.MiB
		}
		return size, nil
	}
	return 256 * unit.MiB, nil
}

func defaultDiskPath() string {
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		candidate := filepath.Join(home, "tmp")
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			return candidate
		}
	}
	return os.TempDir()
}

func diskWarnings(path string, sizeBytes int64, noSync bool) []string {
	var warnings []string
	if noSync {
		warnings = append(warnings, "disk: --no-sync measures buffered writes and may not reflect flushed storage performance")
	}
	if sizeBytes < unit.GiB {
		warnings = append(warnings, "disk: --disk-size below 1G is more sensitive to page cache effects")
	}
	if path == "" {
		path = os.TempDir()
	}
	clean := filepath.Clean(path)
	if clean == "/tmp" || clean == os.TempDir() {
		warnings = append(warnings, "disk: /tmp may be tmpfs or heavily cached; use a real mount point such as /var/tmp or /mnt/data for storage measurements")
	}
	if fsType, ok := system.FilesystemForPath(clean); ok && isMemoryBackedFS(fsType) {
		warnings = append(warnings, fmt.Sprintf("disk: %s is on %s, a memory-backed filesystem; results do not represent a physical disk", clean, fsType))
	}
	return warnings
}

func isMemoryBackedFS(fsType string) bool {
	switch fsType {
	case "tmpfs", "ramfs", "devtmpfs":
		return true
	default:
		return false
	}
}
