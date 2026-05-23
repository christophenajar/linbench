//go:build !linux

package benchmark

import "os"

func fadviseDontNeed(file *os.File) {}
