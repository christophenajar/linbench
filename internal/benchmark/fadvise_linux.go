//go:build linux

package benchmark

import (
	"os"
	"syscall"
)

const posixFadvDontNeed = 4

func fadviseDontNeed(file *os.File) {
	_, _, _ = syscall.Syscall6(syscall.SYS_FADVISE64, file.Fd(), 0, 0, posixFadvDontNeed, 0, 0)
}
