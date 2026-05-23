package main

import "testing"

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
