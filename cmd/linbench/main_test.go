package main

import (
	"testing"

	"linbench/internal/model"
)

func TestIsMemoryBackedFS(t *testing.T) {
	for _, fsType := range []string{"tmpfs", "ramfs", "devtmpfs"} {
		if !isMemoryBackedFS(fsType) {
			t.Fatalf("expected %s to be memory-backed", fsType)
		}
	}
	if isMemoryBackedFS("ext4") {
		t.Fatal("did not expect ext4 to be memory-backed")
	}
}

func TestNUMARecommendations(t *testing.T) {
	recommendations := numaRecommendations(model.NUMAInfo{
		NodeCount: 2,
		Nodes: []model.NUMANodeInfo{
			{ID: 0},
			{ID: 1},
		},
	}, "5s", "1G")
	if len(recommendations) != 2 {
		t.Fatalf("len = %d, want 2", len(recommendations))
	}
	if recommendations[0] != "numactl --cpunodebind=0 --membind=0 ./bin/linbench-linux-amd64 --ram --cache --duration 5s --memory-size 1G" {
		t.Fatalf("local recommendation = %q", recommendations[0])
	}
	if recommendations[1] != "numactl --cpunodebind=0 --membind=1 ./bin/linbench-linux-amd64 --ram --cache --duration 5s --memory-size 1G" {
		t.Fatalf("remote recommendation = %q", recommendations[1])
	}
}
