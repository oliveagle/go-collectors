package collectors

import (
	"testing"
)

func Test_c_network_windows(t *testing.T) {

	// TODO: MSFT_NetLbfoTeamNic only supported above server 2012
	// detect windows version
	// wmic os get caption

	// md, err := c_network_windows()
	// winNetworkInit()
	// md, err := c_network_team_windows()
	// if err != nil {
	// 	t.Error("error:", err)
	// }
	// t.Logf("md count: %d", len(md))

	// isOk := false
	// t.Log("isOk", isOk)
	// for idx := range md {
	// 	t.Log(md[idx])
	// 	// if md[idx].Metric == "win.proc.cpu" && md[idx].Value.(uint64) > 0 {
	// 	// 	isOk = true
	// 	// }
	// }
	// // if !isOk {
	// // 	t.Error("isOk not ok")
	// // }

	// // t.Log(collectors)
	// t.Error("hhh")
}
