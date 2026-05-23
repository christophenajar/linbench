package unit

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	KiB int64 = 1024
	MiB int64 = KiB * 1024
	GiB int64 = MiB * 1024
	TiB int64 = GiB * 1024

	MB float64 = 1_000_000
	GB float64 = 1_000_000_000
)

func ParseSize(s string) (int64, error) {
	original := s
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size")
	}

	multiplier := int64(1)
	last := s[len(s)-1]
	if last < '0' || last > '9' {
		switch strings.ToUpper(string(last)) {
		case "K":
			multiplier = KiB
		case "M":
			multiplier = MiB
		case "G":
			multiplier = GiB
		case "T":
			multiplier = TiB
		case "B":
			multiplier = 1
		default:
			return 0, fmt.Errorf("invalid size suffix %q", last)
		}
		s = strings.TrimSpace(s[:len(s)-1])
	}

	value, err := strconv.ParseFloat(s, 64)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("invalid size %q", original)
	}
	return int64(value * float64(multiplier)), nil
}
