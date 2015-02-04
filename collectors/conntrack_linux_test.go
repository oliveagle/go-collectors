package collectors

import (
	"testing"
)

func Test_c_conntrack_linux(t *testing.T) {
	md, err := c_conntrack_linux()
	if err != nil {
		t.Error("c_conntrack_linux error:", err)
	}
	t.Logf("md count: %d", len(md))
	if len(md) != 3 {
		t.Error("md count incorrect")
	}

	// for idx := range md {
	// 	t.Log(md[idx])
	// }
	// t.Error("hh")
}
