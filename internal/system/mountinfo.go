package system

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type MountInfo struct {
	MountPoint string
	FSType     string
}

func FilesystemForPath(path string) (string, bool) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", false
	}
	defer f.Close()

	abs, err := filepath.Abs(path)
	if err != nil {
		abs = filepath.Clean(path)
	}
	mounts := ParseMountInfo(f)
	var best MountInfo
	for _, mount := range mounts {
		if mount.MountPoint == "" {
			continue
		}
		if isPathOnMount(abs, mount.MountPoint) && len(mount.MountPoint) > len(best.MountPoint) {
			best = mount
		}
	}
	if best.FSType == "" {
		return "", false
	}
	return best.FSType, true
}

func ParseMountInfo(r io.Reader) []MountInfo {
	var mounts []MountInfo
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " - ")
		if len(parts) != 2 {
			continue
		}
		left := strings.Fields(parts[0])
		right := strings.Fields(parts[1])
		if len(left) < 5 || len(right) < 1 {
			continue
		}
		mounts = append(mounts, MountInfo{
			MountPoint: unescapeMountPath(left[4]),
			FSType:     right[0],
		})
	}
	return mounts
}

func isPathOnMount(path, mountPoint string) bool {
	path = filepath.Clean(path)
	mountPoint = filepath.Clean(mountPoint)
	if mountPoint == "/" {
		return strings.HasPrefix(path, "/")
	}
	return path == mountPoint || strings.HasPrefix(path, mountPoint+"/")
}

func unescapeMountPath(path string) string {
	replacer := strings.NewReplacer(
		`\040`, " ",
		`\011`, "\t",
		`\012`, "\n",
		`\134`, `\`,
	)
	return replacer.Replace(path)
}
