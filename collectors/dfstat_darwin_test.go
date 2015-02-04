package collectors

import (
	"testing"
)

func Test_c_dfstat_darwin(t *testing.T) {
	md, err := c_dfstat_darwin()
	t.Log(err)
	t.Log(len(md))

	//
	isOk := false
	for idx := range md {
		t.Log(md[idx])
		if md[idx].Metric == "darwin.disk.fs.total" && md[idx].Value.(int) > 0 {
			isOk = true
		}
	}

	if !isOk {
		t.Error("isOk not ok")
	}

	// t.Error("hh")
}
