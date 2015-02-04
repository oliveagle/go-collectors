package collectors

import (
	"fmt"
	"math/big"
	// "time"

	"github.com/oliveagle/go-collectors/datapoint"
	"github.com/oliveagle/go-collectors/metadata"
)

const (
	ciscoCPU     = ".1.3.6.1.4.1.9.9.109.1.1.1.1.6"
	ciscoMemFree = ".1.3.6.1.4.1.9.9.48.1.1.1.6"
	ciscoMemName = ".1.3.6.1.4.1.9.9.48.1.1.1.2"
	ciscoMemUsed = ".1.3.6.1.4.1.9.9.48.1.1.1.5"
)

func c_snmp_cisco(community, host string) (datapoint.MultiDataPoint, error) {
	var md datapoint.MultiDataPoint
	var v *big.Int
	var err error
	if v, err = snmp_oid(host, community, ciscoCPU); err == nil {
	} else if v, err = snmp_oid(host, community, ciscoCPU+".1"); err == nil {
	} else {
		return nil, err
	}
	Add(&md, "cisco.cpu", v.String(), datapoint.TagSet{"host": host}, metadata.Gauge, metadata.Pct, "The overall CPU busy percentage in the last five-second period.")
	names, err := snmp_subtree(host, community, ciscoMemName)
	if err != nil {
		return nil, err
	}
	used, err := snmp_subtree(host, community, ciscoMemUsed)
	if err != nil {
		return nil, err
	}
	free, err := snmp_subtree(host, community, ciscoMemFree)
	if err != nil {
		return nil, err
	}
	for id, name := range names {
		n := fmt.Sprintf("%s", name)
		u, present := used[id]
		if !present {
			continue
		}
		f, present := free[id]
		if !present {
			continue
		}
		Add(&md, "cisco.mem.used", u, datapoint.TagSet{"host": host, "name": n}, metadata.Unknown, metadata.None, "")
		Add(&md, "cisco.mem.free", f, datapoint.TagSet{"host": host, "name": n}, metadata.Unknown, metadata.None, "")
	}
	return md, nil
}
