package collectors

import (
	"testing"
)

func Test_c_simple_mem_windows(t *testing.T) {
	md, err := c_simple_mem_windows()
	if err != nil {
		t.Error("error:", err)
	}
	t.Logf("md count: %d", len(md))

	isOk := false
	t.Log("isOk", isOk)
	for idx := range md {
		t.Log(md[idx])
		if md[idx].Metric == "os.mem.total" && md[idx].Value.(uint64) > 0 {
			isOk = true
		}
	}
	if !isOk {
		t.Error("isOk not ok")
	}

	// t.Log(collectors)
	// t.Error("hhh")
}
