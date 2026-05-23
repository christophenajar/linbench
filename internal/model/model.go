package model

type BenchmarkResult struct {
	System SystemInfo   `json:"system"`
	Memory MemoryResult `json:"memory"`
	Cache  CacheResult  `json:"cache"`
	CPU    CPUResult    `json:"cpu"`
	Disk   DiskResult   `json:"disk"`
	Errors []string     `json:"errors,omitempty"`
}

type SystemInfo struct {
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	Hostname   string `json:"hostname"`
	Kernel     string `json:"kernel"`
	CPUModel   string `json:"cpu_model"`
	CPUCores   int    `json:"cpu_cores"`
	CPUThreads int    `json:"cpu_threads"`
}

type MemoryResult struct {
	ReadMBS   float64 `json:"read_mb_s"`
	WriteMBS  float64 `json:"write_mb_s"`
	CopyMBS   float64 `json:"copy_mb_s"`
	LatencyNS float64 `json:"latency_ns"`
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

type Sections struct {
	Memory bool
	Cache  bool
	CPU    bool
	Disk   bool
}
