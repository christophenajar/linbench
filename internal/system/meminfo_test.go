package system

import (
	"strings"
	"testing"
)

func TestParseMemInfo(t *testing.T) {
	info := ParseMemInfo(strings.NewReader(`MemTotal:       16384 kB
MemAvailable:    8192 kB
`))
	if info.TotalBytes != 16384*1024 {
		t.Fatalf("TotalBytes = %d", info.TotalBytes)
	}
	if info.AvailableBytes != 8192*1024 {
		t.Fatalf("AvailableBytes = %d", info.AvailableBytes)
	}
}
