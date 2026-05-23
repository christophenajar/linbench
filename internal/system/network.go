package system

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"linbench/internal/model"
)

var perftestTools = []string{
	"ib_write_bw",
	"ib_read_bw",
	"ib_send_bw",
	"ib_write_lat",
	"ib_read_lat",
	"ib_send_lat",
}

func ReadNetworkInfo(includeAll bool) model.NetworkResult {
	result := model.NetworkResult{
		ProcessCPUs:   readProcessCPUAffinity(),
		Scope:         "mellanox-rdma",
		PerftestTools: detectPerftestTools(),
	}
	if includeAll {
		result.Scope = "all"
	}

	entries, err := os.ReadDir("/sys/class/net")
	if err != nil {
		return result
	}
	for _, entry := range entries {
		name := entry.Name()
		if name == "lo" {
			continue
		}
		devicePath := filepath.Join("/sys/class/net", name, "device")
		deviceReal, err := filepath.EvalSymlinks(devicePath)
		if err != nil {
			continue
		}
		iface := readNetworkInterface(name, devicePath, deviceReal, result.ProcessCPUs)
		result.Interfaces = append(result.Interfaces, iface)
	}
	reconcileBondedRDMA(result.Interfaces)
	for i := range result.Interfaces {
		result.Interfaces[i].Warnings = networkWarnings(result.Interfaces[i], result.ProcessCPUs)
	}
	sort.Slice(result.Interfaces, func(i, j int) bool {
		return result.Interfaces[i].Name < result.Interfaces[j].Name
	})
	if !includeAll {
		result.Interfaces, result.HiddenInterfaces = filterRelevantNetworkInterfaces(result.Interfaces)
	}
	for _, iface := range result.Interfaces {
		result.Recommendations = append(result.Recommendations, ifaceRecommendations(iface)...)
	}
	result.Recommendations = uniqueStrings(result.Recommendations)
	return result
}

func readNetworkInterface(name string, devicePath string, deviceReal string, processCPUs string) model.NetworkInterface {
	pciAddress := filepath.Base(deviceReal)
	iface := model.NetworkInterface{
		Name:             name,
		MAC:              readTrimmed(filepath.Join("/sys/class/net", name, "address")),
		OperState:        readTrimmed(filepath.Join("/sys/class/net", name, "operstate")),
		MTU:              readIntFromFile(filepath.Join("/sys/class/net", name, "mtu")),
		SpeedMbps:        readIntFromFile(filepath.Join("/sys/class/net", name, "speed")),
		BondMaster:       readBondMaster(name),
		PCIAddress:       pciAddress,
		VendorID:         readTrimmed(filepath.Join(devicePath, "vendor")),
		DeviceID:         readTrimmed(filepath.Join(devicePath, "device")),
		NUMANode:         readIntFromFile(filepath.Join(devicePath, "numa_node")),
		LocalCPUs:        readTrimmed(filepath.Join(devicePath, "local_cpulist")),
		PCIeCurrentSpeed: readTrimmed(filepath.Join(devicePath, "current_link_speed")),
		PCIeCurrentWidth: readTrimmed(filepath.Join(devicePath, "current_link_width")),
		PCIeMaxSpeed:     readTrimmed(filepath.Join(devicePath, "max_link_speed")),
		PCIeMaxWidth:     readTrimmed(filepath.Join(devicePath, "max_link_width")),
	}
	iface.Driver = driverName(devicePath)
	iface.IsMellanox = strings.EqualFold(iface.VendorID, "0x15b3") || strings.Contains(strings.ToLower(iface.Driver), "mlx5")
	iface.RDMADevice, iface.RDMAAvailable, iface.RoCEAvailable = rdmaForPCI(deviceReal)
	iface.IRQs = readIRQs(devicePath)
	return iface
}

func filterRelevantNetworkInterfaces(ifaces []model.NetworkInterface) ([]model.NetworkInterface, int) {
	var filtered []model.NetworkInterface
	for _, iface := range ifaces {
		if iface.IsMellanox || iface.RDMAAvailable || iface.RoCEAvailable {
			filtered = append(filtered, iface)
		}
	}
	if len(filtered) == 0 {
		return ifaces, 0
	}
	return filtered, len(ifaces) - len(filtered)
}

func reconcileBondedRDMA(ifaces []model.NetworkInterface) {
	bondRDMA := map[string]model.NetworkInterface{}
	for _, iface := range ifaces {
		if iface.BondMaster != "" && iface.RDMAAvailable {
			bondRDMA[iface.BondMaster] = iface
		}
	}
	for i := range ifaces {
		rdma, ok := bondRDMA[ifaces[i].BondMaster]
		if !ok || !ifaces[i].IsMellanox || ifaces[i].RDMAAvailable {
			continue
		}
		ifaces[i].RDMADevice = rdma.RDMADevice
		ifaces[i].RDMAAvailable = true
		ifaces[i].RoCEAvailable = rdma.RoCEAvailable
	}
}

func networkWarnings(iface model.NetworkInterface, processCPUs string) []string {
	var warnings []string
	if iface.IsMellanox && !iface.RDMAAvailable {
		warnings = append(warnings, "RDMA device is missing for Mellanox/NVIDIA interface")
	}
	if iface.RDMAAvailable && !iface.RoCEAvailable {
		warnings = append(warnings, "RDMA device detected but RoCE GID type was not found")
	}
	if iface.SpeedMbps > 0 && iface.SpeedMbps < 25000 && iface.IsMellanox {
		warnings = append(warnings, "link speed is below 25Gb/s")
	}
	if iface.OperState != "" && iface.OperState != "up" {
		warnings = append(warnings, "network link is not up")
	}
	if iface.RoCEAvailable && iface.MTU > 0 && iface.MTU < 9000 {
		warnings = append(warnings, "MTU is below 9000; jumbo frames are usually expected for RoCE")
	}
	if iface.PCIeCurrentSpeed != "" && iface.PCIeMaxSpeed != "" && iface.PCIeCurrentSpeed != iface.PCIeMaxSpeed {
		warnings = append(warnings, "PCIe current link speed is below max link speed")
	}
	if iface.PCIeCurrentWidth != "" && iface.PCIeMaxWidth != "" && iface.PCIeCurrentWidth != iface.PCIeMaxWidth {
		warnings = append(warnings, "PCIe current link width is below max link width")
	}
	if iface.LocalCPUs != "" && CPUListContainsRemote(processCPUs, iface.LocalCPUs) {
		warnings = append(warnings, "process CPU affinity includes CPUs outside NIC-local NUMA node")
	}
	remoteIRQs := 0
	for _, irq := range iface.IRQs {
		if irq.Affinity != "" && iface.LocalCPUs != "" && CPUListContainsRemote(irq.Affinity, iface.LocalCPUs) {
			remoteIRQs++
		}
	}
	if remoteIRQs > 0 {
		warnings = append(warnings, "one or more NIC IRQ affinities include CPUs outside NIC-local NUMA node")
	}
	return warnings
}

func ifaceRecommendations(iface model.NetworkInterface) []string {
	if iface.NUMANode < 0 || iface.LocalCPUs == "" || !iface.IsMellanox {
		return nil
	}
	return []string{
		"numactl --cpunodebind=" + strconv.Itoa(iface.NUMANode) + " --membind=" + strconv.Itoa(iface.NUMANode) + " ./bin/linbench-linux-amd64 --network",
	}
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	var unique []string
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}

func readProcessCPUAffinity() string {
	f, err := os.Open("/proc/self/status")
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Cpus_allowed_list:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Cpus_allowed_list:"))
		}
	}
	return ""
}

func detectPerftestTools() map[string]bool {
	tools := map[string]bool{}
	for _, tool := range perftestTools {
		_, err := exec.LookPath(tool)
		tools[tool] = err == nil
	}
	return tools
}

func driverName(devicePath string) string {
	driverPath, err := filepath.EvalSymlinks(filepath.Join(devicePath, "driver"))
	if err != nil {
		return ""
	}
	return filepath.Base(driverPath)
}

func readBondMaster(name string) string {
	master, err := filepath.EvalSymlinks(filepath.Join("/sys/class/net", name, "master"))
	if err != nil {
		return ""
	}
	return filepath.Base(master)
}

func readIntFromFile(path string) int {
	value, err := strconv.Atoi(readTrimmed(path))
	if err != nil {
		return 0
	}
	return value
}

func rdmaForPCI(deviceReal string) (name string, available bool, roce bool) {
	rdmaDirs, _ := filepath.Glob(filepath.Join(deviceReal, "infiniband", "*"))
	if len(rdmaDirs) == 0 {
		rdmaDirs, _ = filepath.Glob("/sys/class/infiniband/*")
	}
	for _, dir := range rdmaDirs {
		rdmaName := filepath.Base(dir)
		rdmaDeviceReal, err := filepath.EvalSymlinks(filepath.Join("/sys/class/infiniband", rdmaName, "device"))
		if err != nil && strings.HasPrefix(dir, deviceReal) {
			rdmaDeviceReal = deviceReal
		}
		if rdmaDeviceReal != deviceReal {
			continue
		}
		return rdmaName, true, roceAvailable(rdmaName)
	}
	return "", false, false
}

func roceAvailable(rdmaName string) bool {
	typeFiles, _ := filepath.Glob(filepath.Join("/sys/class/infiniband", rdmaName, "ports", "*", "gid_attrs", "types", "*"))
	for _, path := range typeFiles {
		value := strings.ToLower(readTrimmed(path))
		if strings.Contains(value, "roce") {
			return true
		}
	}
	return false
}

func readIRQs(devicePath string) []model.IRQInfo {
	irqDirs, err := os.ReadDir(filepath.Join(devicePath, "msi_irqs"))
	if err != nil {
		return nil
	}
	interruptNames := readInterruptNames()
	irqs := make([]model.IRQInfo, 0, len(irqDirs))
	for _, entry := range irqDirs {
		irq := entry.Name()
		irqs = append(irqs, model.IRQInfo{
			IRQ:      irq,
			Name:     interruptNames[irq],
			Affinity: readTrimmed(filepath.Join("/proc/irq", irq, "smp_affinity_list")),
		})
	}
	sort.Slice(irqs, func(i, j int) bool {
		left, _ := strconv.Atoi(irqs[i].IRQ)
		right, _ := strconv.Atoi(irqs[j].IRQ)
		return left < right
	})
	return irqs
}

func readInterruptNames() map[string]string {
	names := map[string]string{}
	f, err := os.Open("/proc/interrupts")
	if err != nil {
		return names
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		irq := strings.TrimSpace(parts[0])
		fields := strings.Fields(parts[1])
		if len(fields) > 0 {
			names[irq] = fields[len(fields)-1]
		}
	}
	return names
}
