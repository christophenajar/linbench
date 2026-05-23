package unit

import "testing"

func TestParseSize(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"512M", 512 * MiB},
		{"1G", GiB},
		{"4096K", 4096 * KiB},
		{"123", 123},
		{"1.5G", GiB + GiB/2},
	}

	for _, tt := range tests {
		got, err := ParseSize(tt.input)
		if err != nil {
			t.Fatalf("ParseSize(%q) error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("ParseSize(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseSizeRejectsInvalidInput(t *testing.T) {
	if _, err := ParseSize("10X"); err == nil {
		t.Fatal("expected error for invalid suffix")
	}
}
