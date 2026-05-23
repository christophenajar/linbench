package system

import (
	"os"
	"strings"
)

func ReadDMIField(name string) string {
	value, err := os.ReadFile("/sys/class/dmi/id/" + name)
	if err != nil {
		return "unknown"
	}
	value = []byte(strings.TrimSpace(string(value)))
	if len(value) == 0 {
		return "unknown"
	}
	return string(value)
}
