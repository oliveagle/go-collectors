// +build linux
package collectors

import (
	"bosun.org/slog"
	"testing"
)

func Test_c_procstats_linux(t *testing.T) {
	md, err := c_procstats_linux()
	t.Log(err)
	t.Log(len(md))
	if len(md) < 100 {
		t.Error("md count below 100")
	}

	slog.Info("hahah")

	isOsMemTotalOk := false
	for idx := range md {
		t.Log(md[idx])
		if md[idx].Metric == osMemTotal && md[idx].Value.(int) > 0 {
			isOsMemTotalOk = true
		}
	}
	if !isOsMemTotalOk {
		t.Error("os mem total not ok")
	}

	t.Log(collectors)
	// t.Error("hhh")
}
