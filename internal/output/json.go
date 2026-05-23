package output

import (
	"encoding/json"

	"linbench/internal/model"
)

func FormatJSON(result model.BenchmarkResult) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}
