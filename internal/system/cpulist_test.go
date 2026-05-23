package system

import "testing"

func TestParseCPUList(t *testing.T) {
	cpus := ParseCPUList("0,2,4-6")
	for _, cpu := range []int{0, 2, 4, 5, 6} {
		if !cpus[cpu] {
			t.Fatalf("missing CPU %d", cpu)
		}
	}
	if cpus[1] {
		t.Fatal("unexpected CPU 1")
	}
}

func TestCPUListContainsRemote(t *testing.T) {
	if !CPUListContainsRemote("0-3", "0-1") {
		t.Fatal("expected remote CPU")
	}
	if CPUListContainsRemote("0,1", "0-3") {
		t.Fatal("did not expect remote CPU")
	}
}

func TestFormatCPUSet(t *testing.T) {
	got := FormatCPUSet(map[int]bool{0: true, 1: true, 2: true, 4: true, 6: true, 7: true})
	if got != "0-2,4,6-7" {
		t.Fatalf("FormatCPUSet = %q", got)
	}
}
