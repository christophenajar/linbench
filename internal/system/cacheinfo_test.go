package system

import (
	"testing"

	"linbench/internal/unit"
)

func TestParseCacheSize(t *testing.T) {
	tests := map[string]int64{
		"32K": 32 * unit.KiB,
		"1M":  unit.MiB,
		"96K": 96 * unit.KiB,
	}
	for input, want := range tests {
		got, err := ParseCacheSize(input)
		if err != nil {
			t.Fatalf("ParseCacheSize(%q) error: %v", input, err)
		}
		if got != want {
			t.Fatalf("ParseCacheSize(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestCacheSizesByLevelPrefersL1Data(t *testing.T) {
	sizes := CacheSizesByLevel([]CacheInfo{
		{Level: 1, Type: "Instruction", SizeBytes: 64 * unit.KiB},
		{Level: 1, Type: "Data", SizeBytes: 32 * unit.KiB},
		{Level: 2, Type: "Unified", SizeBytes: 512 * unit.KiB},
	})
	if sizes[1] != 32*unit.KiB {
		t.Fatalf("L1 size = %d", sizes[1])
	}
	if sizes[2] != 512*unit.KiB {
		t.Fatalf("L2 size = %d", sizes[2])
	}
}
