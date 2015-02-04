package collectors

import (
	"testing"
)

func Test_c_iostat_linux(t *testing.T) {
	md, err := c_iostat_linux()
	//
	// if err != nil {
	// 	t.Error(err)
	// }
	t.Log(len(md), err)

	isOk := false
	for idx := range md {
		t.Log(md[idx])
	}
	t.Log(isOk)
	// if !isOk {
	// 	t.Error("is not ok")
	// }
	// t.Error("hhh")
	// linux.disk.rem.read_requests
}

func Test_c_dfstat_blocks_linux(t *testing.T) {
	md, err := c_dfstat_blocks_linux()
	if err != nil {
		t.Error(err)
	}
	t.Log(len(md))

	isOk := false
	for idx := range md {
		t.Log(md[idx])
	}
	t.Log(isOk)
	// if !isOk {
	// 	t.Error("is not ok")
	// }
	// t.Error("hhh")

}

func Test_c_dfstat_inodes_linux(t *testing.T) {
	md, err := c_dfstat_inodes_linux()
	if err != nil {
		t.Error(err)
	}
	t.Log(len(md))

	isOk := false
	for idx := range md {
		t.Log(md[idx])
	}
	t.Log(isOk)
	// if !isOk {
	// 	t.Error("is not ok")
	// }
	// t.Error("hhh")

}
