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

func TestFilterRelevantNetworkInterfaces(t *testing.T) {
	filtered, hidden := filterRelevantNetworkInterfaces([]model.NetworkInterface{
		{Name: "eno1", Driver: "tg3"},
		{Name: "enp4s0f0np0", Driver: "mlx5_core", IsMellanox: true},
	})
	if len(filtered) != 1 || filtered[0].Name != "enp4s0f0np0" {
		t.Fatalf("filtered = %#v", filtered)
	}
	if hidden != 1 {
		t.Fatalf("hidden = %d", hidden)
	}
}

func TestReconcileBondedRDMA(t *testing.T) {
	ifaces := []model.NetworkInterface{
		{Name: "enp4s0f0np0", IsMellanox: true, BondMaster: "bond0", RDMAAvailable: true, RDMADevice: "mlx5_bond_0", RoCEAvailable: true},
		{Name: "enp4s0f1np1", IsMellanox: true, BondMaster: "bond0"},
	}
	reconcileBondedRDMA(ifaces)
	if !ifaces[1].RDMAAvailable || ifaces[1].RDMADevice != "mlx5_bond_0" || !ifaces[1].RoCEAvailable {
		t.Fatalf("bonded RDMA was not propagated: %#v", ifaces[1])
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
