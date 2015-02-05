package collectors

import (
	"testing"
)

func Test_c_cpu_windows(t *testing.T) {
	md, err := c_cpu_windows()
	if err != nil {
		t.Error("error:", err)
	}
	t.Logf("md count: %d", len(md))

	isOsCPUOk := false
	for idx := range md {
		// t.Log(md[idx])
		if md[idx].Metric == osCPU && md[idx].Value.(uint64) > 0 {
			isOsCPUOk = true
		}
	}
	if !isOsCPUOk {
		t.Error("isOsCPUOk not ok")
	}

	// t.Log(collectors)
	// t.Error("hhh")
}
