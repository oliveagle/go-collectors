// +build linux
package collectors

import (
	"testing"
)

func Test_puppet_linux(t *testing.T) {
	md, err := puppet_linux()
	t.Log(err)
	t.Log(len(md))

	isOk := false
	t.Log("isOk: ", isOk)
	for idx := range md {
		t.Log(md[idx])
	}
	// if !isOk {
	// 	t.Error("is not ok")
	// }

	// t.Log(collectors)
	// t.Error("hhh")
}
