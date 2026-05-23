package output

import (
	"encoding/json"
	"strings"
	"testing"

	"linbench/internal/model"
)

func TestFormatJSON(t *testing.T) {
	result := sampleResult()
	data, err := FormatJSON(result)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}
	var decoded model.BenchmarkResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if decoded.System.CPUModel != "Example CPU" {
		t.Fatalf("decoded CPU model = %q", decoded.System.CPUModel)
	}
}

func TestFormatText(t *testing.T) {
	result := sampleResult()
	result.Warnings = []string{"disk: example warning"}
	text := FormatText(result, model.Sections{Memory: true, Cache: true, CPU: true, Disk: true})
	for _, want := range []string{"Memory", "L1 Cache", "CPU", "Disk", "Warnings", "Note:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("text output missing %q:\n%s", want, text)
		}
	}
}

func sampleResult() model.BenchmarkResult {
	return model.BenchmarkResult{
		System: model.SystemInfo{
			OS:         "linux",
			Arch:       "amd64",
			Hostname:   "host",
			Kernel:     "6.8.0",
			CPUModel:   "Example CPU",
			CPUCores:   8,
			CPUThreads: 16,
		},
		Memory: model.MemoryResult{ReadMBS: 1, WriteMBS: 2, CopyMBS: 3, LatencyNS: 4},
		Cache: model.CacheResult{
			L1: model.CacheLevelResult{SizeBytes: 1, ReadGBS: 2, WriteGBS: 3, CopyGBS: 4, LatencyNS: 5},
			L2: model.CacheLevelResult{SizeBytes: 1, ReadGBS: 2, WriteGBS: 3, CopyGBS: 4, LatencyNS: 5},
			L3: model.CacheLevelResult{SizeBytes: 1, ReadGBS: 2, WriteGBS: 3, CopyGBS: 4, LatencyNS: 5},
		},
		CPU:  model.CPUResult{IntegerScore: 1, FloatScore: 2, SHA256MBS: 3, CompressionMBS: 4},
		Disk: model.DiskResult{Path: "/tmp", SequentialReadMBS: 1, SequentialWriteMBS: 2, RandomReadIOPS: 3, RandomWriteIOPS: 4, LatencyMS: 5},
	}
}
