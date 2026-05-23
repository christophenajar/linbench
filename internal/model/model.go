package model

type BenchmarkResult struct {
	System          SystemInfo    `json:"system"`
	BenchmarkEngine string        `json:"benchmark_engine"`
	NUMA            NUMAInfo      `json:"numa"`
	Memory          MemoryResult  `json:"memory"`
	Cache           CacheResult   `json:"cache"`
	CPU             CPUResult     `json:"cpu"`
	Disk            DiskResult    `json:"disk"`
	Network         NetworkResult `json:"network"`
	Warnings        []string      `json:"warnings,omitempty"`
	Errors          []string      `json:"errors,omitempty"`
}

type SystemInfo struct {
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	Hostname   string `json:"hostname"`
	Kernel     string `json:"kernel"`
	CPUModel   string `json:"cpu_model"`
	CPUSockets int    `json:"cpu_sockets"`
	CPUCores   int    `json:"cpu_cores"`
	CPUThreads int    `json:"cpu_threads"`
}

type MemoryResult struct {
	ReadMBS   float64 `json:"read_mb_s"`
	WriteMBS  float64 `json:"write_mb_s"`
	CopyMBS   float64 `json:"copy_mb_s"`
	LatencyNS float64 `json:"latency_ns"`
}

type NUMAInfo struct {
	Available       bool           `json:"available"`
	NodeCount       int            `json:"node_count"`
	Nodes           []NUMANodeInfo `json:"nodes,omitempty"`
	Recommendations []string       `json:"recommendations,omitempty"`
}

type NUMANodeInfo struct {
	ID            int    `json:"id"`
	CPUs          string `json:"cpus"`
	MemTotalBytes int64  `json:"mem_total_bytes"`
	MemFreeBytes  int64  `json:"mem_free_bytes"`
	Distance      []int  `json:"distance,omitempty"`
}

type CacheLevelResult struct {
	SizeBytes int64   `json:"size_bytes"`
	ReadGBS   float64 `json:"read_gb_s"`
	WriteGBS  float64 `json:"write_gb_s"`
	CopyGBS   float64 `json:"copy_gb_s"`
	LatencyNS float64 `json:"latency_ns"`
}

type CacheResult struct {
	L1 CacheLevelResult `json:"l1"`
	L2 CacheLevelResult `json:"l2"`
	L3 CacheLevelResult `json:"l3"`
}

type CPUResult struct {
	IntegerScore   float64 `json:"integer_score"`
	FloatScore     float64 `json:"float_score"`
	SHA256MBS      float64 `json:"sha256_mb_s"`
	CompressionMBS float64 `json:"compression_mb_s"`
}

type DiskResult struct {
	Path               string  `json:"path"`
	SequentialReadMBS  float64 `json:"sequential_read_mb_s"`
	SequentialWriteMBS float64 `json:"sequential_write_mb_s"`
	RandomReadIOPS     float64 `json:"random_read_iops"`
	RandomWriteIOPS    float64 `json:"random_write_iops"`
	LatencyMS          float64 `json:"latency_ms"`
}

type NetworkResult struct {
	ProcessCPUs      string             `json:"process_cpus"`
	Scope            string             `json:"scope"`
	HiddenInterfaces int                `json:"hidden_interfaces,omitempty"`
	Interfaces       []NetworkInterface `json:"interfaces,omitempty"`
	PerftestTools    map[string]bool    `json:"perftest_tools,omitempty"`
	Recommendations  []string           `json:"recommendations,omitempty"`
}

type NetworkInterface struct {
	Name             string    `json:"name"`
	Driver           string    `json:"driver"`
	MAC              string    `json:"mac"`
	OperState        string    `json:"oper_state"`
	MTU              int       `json:"mtu"`
	SpeedMbps        int       `json:"speed_mbps"`
	BondMaster       string    `json:"bond_master,omitempty"`
	PCIAddress       string    `json:"pci_address"`
	VendorID         string    `json:"vendor_id"`
	DeviceID         string    `json:"device_id"`
	IsMellanox       bool      `json:"is_mellanox"`
	NUMANode         int       `json:"numa_node"`
	LocalCPUs        string    `json:"local_cpus"`
	PCIeCurrentSpeed string    `json:"pcie_current_speed"`
	PCIeCurrentWidth string    `json:"pcie_current_width"`
	PCIeMaxSpeed     string    `json:"pcie_max_speed"`
	PCIeMaxWidth     string    `json:"pcie_max_width"`
	RDMADevice       string    `json:"rdma_device,omitempty"`
	RDMAAvailable    bool      `json:"rdma_available"`
	RoCEAvailable    bool      `json:"roce_available"`
	IRQs             []IRQInfo `json:"irqs,omitempty"`
	Warnings         []string  `json:"warnings,omitempty"`
}

type IRQInfo struct {
	IRQ      string `json:"irq"`
	Name     string `json:"name,omitempty"`
	Affinity string `json:"affinity"`
}

type Sections struct {
	Memory  bool
	Cache   bool
	CPU     bool
	Disk    bool
	Network bool
}
