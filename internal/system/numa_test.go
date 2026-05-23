package system

import (
	"strings"
	"testing"
)

func TestParseNodeID(t *testing.T) {
	id, ok := parseNodeID("node12")
	if !ok || id != 12 {
		t.Fatalf("parseNodeID = %d, %v", id, ok)
	}
	if _, ok := parseNodeID("cpu0"); ok {
		t.Fatal("expected cpu0 to be rejected")
	}
}

func TestParseDistanceLine(t *testing.T) {
	got := parseDistanceLine("10 21 35")
	want := []int{10, 21, 35}
	if len(got) != len(want) {
		t.Fatalf("len = %d", len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("distance[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestParseNUMANodeMemInfo(t *testing.T) {
	input := `Node 0 MemTotal:       131911080 kB
Node 0 MemFree:        129466368 kB
`
	total, free := parseNUMANodeMemInfo(strings.NewReader(input))
	if total != 131911080*1024 {
		t.Fatalf("total = %d", total)
	}
	if free != 129466368*1024 {
		t.Fatalf("free = %d", free)
	}
}
