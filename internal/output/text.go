package output

import (
	"fmt"
	"sort"
	"strings"

	"linbench/internal/model"
)

func FormatText(result model.BenchmarkResult, sections model.Sections) string {
	var b strings.Builder
	fmt.Fprintln(&b, "Linux Hardware Benchmark")
	fmt.Fprintln(&b, "========================")
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "System")
	fmt.Fprintf(&b, "  OS        : %s/%s\n", result.System.OS, result.System.Arch)
	fmt.Fprintf(&b, "  Hostname  : %s\n", result.System.Hostname)
	fmt.Fprintf(&b, "  Kernel    : %s\n", result.System.Kernel)
	fmt.Fprintf(&b, "  CPU Model : %s\n", result.System.CPUModel)
	fmt.Fprintf(&b, "  Sockets   : %d\n", result.System.CPUSockets)
	fmt.Fprintf(&b, "  Cores     : %d\n", result.System.CPUCores)
	fmt.Fprintf(&b, "  Threads   : %d\n", result.System.CPUThreads)
	if result.BenchmarkEngine != "" {
		fmt.Fprintf(&b, "  Engine    : %s\n", result.BenchmarkEngine)
	}
	fmt.Fprintln(&b)

	if result.NUMA.Available {
		writeNUMA(&b, result.NUMA)
	}

	if sections.Memory {
		fmt.Fprintln(&b, "Memory")
		fmt.Fprintf(&b, "  Read      : %.0f MB/s\n", result.Memory.ReadMBS)
		fmt.Fprintf(&b, "  Write     : %.0f MB/s\n", result.Memory.WriteMBS)
		fmt.Fprintf(&b, "  Copy      : %.0f MB/s\n", result.Memory.CopyMBS)
		fmt.Fprintf(&b, "  Latency   : %.1f ns\n", result.Memory.LatencyNS)
		fmt.Fprintln(&b)
	}

	if sections.Cache {
		writeCacheLevel(&b, "L1 Cache", result.Cache.L1)
		writeCacheLevel(&b, "L2 Cache", result.Cache.L2)
		writeCacheLevel(&b, "L3 Cache", result.Cache.L3)
	}

	if sections.CPU {
		fmt.Fprintln(&b, "CPU")
		fmt.Fprintf(&b, "  Integer   : %.0f ops/s\n", result.CPU.IntegerScore)
		fmt.Fprintf(&b, "  Float     : %.0f ops/s\n", result.CPU.FloatScore)
		fmt.Fprintf(&b, "  SHA-256   : %.0f MB/s\n", result.CPU.SHA256MBS)
		fmt.Fprintf(&b, "  Gzip      : %.0f MB/s\n", result.CPU.CompressionMBS)
		fmt.Fprintln(&b)
	}

	if sections.Disk {
		fmt.Fprintln(&b, "Disk")
		fmt.Fprintf(&b, "  Path      : %s\n", result.Disk.Path)
		fmt.Fprintf(&b, "  Read      : %.0f MB/s\n", result.Disk.SequentialReadMBS)
		fmt.Fprintf(&b, "  Write     : %.0f MB/s\n", result.Disk.SequentialWriteMBS)
		fmt.Fprintf(&b, "  IOPS Read : %.0f\n", result.Disk.RandomReadIOPS)
		fmt.Fprintf(&b, "  IOPS Write: %.0f\n", result.Disk.RandomWriteIOPS)
		fmt.Fprintf(&b, "  Latency   : %.3f ms\n", result.Disk.LatencyMS)
		fmt.Fprintln(&b)
	}

	if sections.Network {
		writeNetwork(&b, result.Network)
	}

	if len(result.Warnings) > 0 {
		fmt.Fprintln(&b, "Warnings")
		for _, warning := range result.Warnings {
			fmt.Fprintf(&b, "  - %s\n", warning)
		}
		fmt.Fprintln(&b)
	}

	if len(result.Errors) > 0 {
		fmt.Fprintln(&b, "Errors")
		for _, err := range result.Errors {
			fmt.Fprintf(&b, "  - %s\n", err)
		}
		fmt.Fprintln(&b)
	}

	fmt.Fprintln(&b, "Note: results may vary depending on CPU frequency scaling, system load, thermal limits and OS cache.")
	return b.String()
}

func writeCacheLevel(b *strings.Builder, name string, result model.CacheLevelResult) {
	fmt.Fprintln(b, name)
	fmt.Fprintf(b, "  Size      : %d bytes\n", result.SizeBytes)
	fmt.Fprintf(b, "  Read      : %.1f GB/s\n", result.ReadGBS)
	fmt.Fprintf(b, "  Write     : %.1f GB/s\n", result.WriteGBS)
	fmt.Fprintf(b, "  Copy      : %.1f GB/s\n", result.CopyGBS)
	fmt.Fprintf(b, "  Latency   : %.1f ns\n", result.LatencyNS)
	fmt.Fprintln(b)
}

func writeNUMA(b *strings.Builder, result model.NUMAInfo) {
	fmt.Fprintln(b, "NUMA")
	fmt.Fprintf(b, "  Nodes     : %d\n", result.NodeCount)
	for _, node := range result.Nodes {
		fmt.Fprintf(b, "  Node %d\n", node.ID)
		fmt.Fprintf(b, "    CPUs    : %s\n", valueOrUnknown(node.CPUs))
		if node.MemTotalBytes > 0 {
			fmt.Fprintf(b, "    Memory  : %.0f MB total, %.0f MB free\n", bytesToMB(node.MemTotalBytes), bytesToMB(node.MemFreeBytes))
		} else {
			fmt.Fprintln(b, "    Memory  : unknown")
		}
		if len(node.Distance) > 0 {
			fmt.Fprintf(b, "    Distance: %s\n", intsToString(node.Distance))
		}
	}
	if len(result.Recommendations) > 0 {
		fmt.Fprintln(b, "  Recommended")
		for _, recommendation := range result.Recommendations {
			fmt.Fprintf(b, "    %s\n", recommendation)
		}
	}
	fmt.Fprintln(b)
}

func bytesToMB(value int64) float64 {
	return float64(value) / 1_000_000
}

func valueOrUnknown(value string) string {
	if value == "" {
		return "unknown"
	}
	return value
}

func intsToString(values []int) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, fmt.Sprint(value))
	}
	return strings.Join(parts, " ")
}

func writeNetwork(b *strings.Builder, result model.NetworkResult) {
	fmt.Fprintln(b, "Network / RDMA")
	if result.ProcessCPUs != "" {
		fmt.Fprintf(b, "  Process CPUs: %s\n", result.ProcessCPUs)
	}
	if result.Scope != "" {
		fmt.Fprintf(b, "  Scope       : %s\n", result.Scope)
	}
	if result.HiddenInterfaces > 0 {
		fmt.Fprintf(b, "  Hidden      : %d non-RDMA/non-Mellanox interfaces (use --network-all)\n", result.HiddenInterfaces)
	}
	if len(result.Interfaces) == 0 {
		fmt.Fprintln(b, "  Interfaces  : none detected")
		fmt.Fprintln(b)
		return
	}
	for _, iface := range result.Interfaces {
		fmt.Fprintf(b, "  Interface %s\n", iface.Name)
		fmt.Fprintf(b, "    Driver   : %s\n", valueOrUnknown(iface.Driver))
		fmt.Fprintf(b, "    MAC      : %s\n", valueOrUnknown(iface.MAC))
		fmt.Fprintf(b, "    State    : %s\n", valueOrUnknown(iface.OperState))
		if iface.BondMaster != "" {
			fmt.Fprintf(b, "    Bond     : %s\n", iface.BondMaster)
		}
		fmt.Fprintf(b, "    MTU      : %d\n", iface.MTU)
		fmt.Fprintf(b, "    Speed    : %d Mb/s\n", iface.SpeedMbps)
		fmt.Fprintf(b, "    PCI      : %s\n", valueOrUnknown(iface.PCIAddress))
		fmt.Fprintf(b, "    NUMA Node: %d\n", iface.NUMANode)
		fmt.Fprintf(b, "    Local CPU: %s\n", valueOrUnknown(iface.LocalCPUs))
		if iface.PCIeCurrentSpeed != "" || iface.PCIeCurrentWidth != "" {
			fmt.Fprintf(b, "    PCIe     : %s x%s", valueOrUnknown(iface.PCIeCurrentSpeed), valueOrUnknown(iface.PCIeCurrentWidth))
			if iface.PCIeMaxSpeed != "" || iface.PCIeMaxWidth != "" {
				fmt.Fprintf(b, " (max %s x%s)", valueOrUnknown(iface.PCIeMaxSpeed), valueOrUnknown(iface.PCIeMaxWidth))
			}
			fmt.Fprintln(b)
		}
		if iface.RDMAAvailable {
			fmt.Fprintf(b, "    RDMA     : %s\n", iface.RDMADevice)
			fmt.Fprintf(b, "    RoCE     : %t\n", iface.RoCEAvailable)
		} else {
			fmt.Fprintln(b, "    RDMA     : unavailable")
		}
		if len(iface.IRQs) > 0 {
			fmt.Fprintf(b, "    IRQs     : %d\n", len(iface.IRQs))
			limit := len(iface.IRQs)
			if limit > 16 {
				limit = 16
			}
			for i := 0; i < limit; i++ {
				irq := iface.IRQs[i]
				name := irq.Name
				if name == "" {
					name = "unknown"
				}
				fmt.Fprintf(b, "      %s %-24s %s\n", irq.IRQ, name, valueOrUnknown(irq.Affinity))
			}
			if len(iface.IRQs) > limit {
				fmt.Fprintf(b, "      ... %d more IRQs\n", len(iface.IRQs)-limit)
			}
		}
		if len(iface.Warnings) > 0 {
			fmt.Fprintln(b, "    Warnings")
			for _, warning := range iface.Warnings {
				fmt.Fprintf(b, "      - %s\n", warning)
			}
		}
	}
	if len(result.PerftestTools) > 0 {
		fmt.Fprintln(b, "  Perftest")
		for _, tool := range sortedToolNames(result.PerftestTools) {
			fmt.Fprintf(b, "    %-12s: %t\n", tool, result.PerftestTools[tool])
		}
	}
	if len(result.Recommendations) > 0 {
		fmt.Fprintln(b, "  Recommended")
		for _, recommendation := range result.Recommendations {
			fmt.Fprintf(b, "    %s\n", recommendation)
		}
	}
	fmt.Fprintln(b)
}

func sortedToolNames(tools map[string]bool) []string {
	names := make([]string, 0, len(tools))
	for name := range tools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
