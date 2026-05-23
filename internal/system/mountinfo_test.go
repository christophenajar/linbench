package system

import (
	"strings"
	"testing"
)

func TestParseMountInfo(t *testing.T) {
	input := `25 1 0:24 / / rw,relatime - ext4 /dev/sda1 rw
26 25 0:25 / /tmp rw,nosuid,nodev - tmpfs tmpfs rw,size=1024k
27 25 0:26 / /mnt/data\040set rw,relatime - xfs /dev/sdb1 rw
`
	mounts := ParseMountInfo(strings.NewReader(input))
	if len(mounts) != 3 {
		t.Fatalf("len(mounts) = %d, want 3", len(mounts))
	}
	if mounts[1].MountPoint != "/tmp" || mounts[1].FSType != "tmpfs" {
		t.Fatalf("mount[1] = %+v", mounts[1])
	}
	if mounts[2].MountPoint != "/mnt/data set" || mounts[2].FSType != "xfs" {
		t.Fatalf("mount[2] = %+v", mounts[2])
	}
}

func TestIsPathOnMount(t *testing.T) {
	if !isPathOnMount("/tmp/linbench-file", "/tmp") {
		t.Fatal("expected path to be on /tmp")
	}
	if isPathOnMount("/tmp2/file", "/tmp") {
		t.Fatal("did not expect /tmp2/file to be on /tmp")
	}
}
