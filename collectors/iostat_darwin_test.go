package collectors

import (
	"testing"
)

func Test_c_iostat_darwin(t *testing.T) {
	md, err := c_iostat_darwin()
	t.Log(err)
	t.Log(len(md))
	//
	isOk := false
	for idx := range md {
		t.Log(md[idx])
		if md[idx].Metric == "darwin.loadavg_1_min" && md[idx].Value.(float64) > 0 {
			isOk = true
		}
	}

	if !isOk {
		t.Error("isOk not ok")
	}

	// t.Error("hh")
}
