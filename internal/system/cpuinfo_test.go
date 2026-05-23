package system

import (
	"strings"
	"testing"
)

func TestParseCPUInfo(t *testing.T) {
	input := `processor	: 0
physical id	: 0
core id		: 0
cpu cores	: 2
model name	: Example CPU

processor	: 1
physical id	: 0
core id		: 1
cpu cores	: 2
model name	: Example CPU
`
	model, sockets, cores, threads := ParseCPUInfo(strings.NewReader(input))
	if model != "Example CPU" {
		t.Fatalf("model = %q", model)
	}
	if sockets != 1 {
		t.Fatalf("sockets = %d, want 1", sockets)
	}
	if cores != 2 {
		t.Fatalf("cores = %d, want 2", cores)
	}
	if threads != 2 {
		t.Fatalf("threads = %d, want 2", threads)
	}
}
