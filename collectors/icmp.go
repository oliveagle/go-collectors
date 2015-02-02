package collectors

import (
	"fmt"
	"net"
	"time"

	"bosun.org/_third_party/github.com/tatsushid/go-fastping"
	"github.com/oliveagle/go-collectors/datapoint"
	"github.com/oliveagle/go-collectors/metadata"
)

type response struct {
	addr *net.IPAddr
	rtt  time.Duration
}

// ICMP registers an ICMP collector a given host.
func ICMP(host string) {
	collectors = append(collectors, &IntervalCollector{
		F: func() (datapoint.MultiDataPoint, error) {
			return c_icmp(host)
		},
		name: fmt.Sprintf("icmp-%s", host),
	})
}

func c_icmp(host string) (datapoint.MultiDataPoint, error) {
	var md datapoint.MultiDataPoint
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", host)
	if err != nil {
		return nil, err
	}
	p.AddIPAddr(ra)
	p.MaxRTT = time.Second * 5
	timeout := 1
	p.OnRecv = func(addr *net.IPAddr, t time.Duration) {
		Add(&md, "ping.rtt", float64(t)/float64(time.Millisecond), datapoint.TagSet{"dst_host": host}, metadata.Unknown, metadata.None, "")
		timeout = 0
	}
	if err := p.Run(); err != nil {
		return nil, err
	}
	Add(&md, "ping.timeout", timeout, datapoint.TagSet{"dst_host": host}, metadata.Unknown, metadata.None, "")
	return md, nil
}
