package system

import (
	"testing"

	"linbench/internal/model"
)

func TestNetworkWarnings(t *testing.T) {
	warnings := networkWarnings(model.NetworkInterface{
		Name:          "enp4s0f0np0",
		IsMellanox:    true,
		RDMAAvailable: false,
		OperState:     "up",
		MTU:           1500,
		SpeedMbps:     10000,
		LocalCPUs:     "0-3",
		IRQs: []model.IRQInfo{
			{IRQ: "74", Affinity: "4-7"},
		},
	}, "0-7")
	for _, want := range []string{
		"RDMA device is missing for Mellanox/NVIDIA interface",
		"link speed is below 25Gb/s",
		"process CPU affinity includes CPUs outside NIC-local NUMA node",
		"one or more NIC IRQ affinities include CPUs outside NIC-local NUMA node",
	} {
		if !containsString(warnings, want) {
			t.Fatalf("missing warning %q in %#v", want, warnings)
		}
	}
}

func TestIfaceRecommendations(t *testing.T) {
	recommendations := ifaceRecommendations(model.NetworkInterface{
		IsMellanox: true,
		NUMANode:   0,
		LocalCPUs:  "0-3",
	})
	if len(recommendations) != 1 {
		t.Fatalf("len = %d", len(recommendations))
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
