package collectors

import (
	"testing"
)

func Test_c_diskspace_windows(t *testing.T) {
	md, err := c_diskspace_windows()
	if err != nil {
		t.Error("error:", err)
	}
	t.Logf("md count: %d", len(md))

	isOk := false
	t.Log("isOk", isOk)
	for idx := range md {
		t.Log(md[idx])
	}
	// if !isOk {
	// 	t.Error("isOsCPUOk not ok")
	// }

	// t.Log(collectors)
	// t.Error("hhh")
}
