package collectors

import (
	"fmt"
	"strings"
	"time"

	"github.com/oliveagle/go-collectors/datapoint"
	"github.com/oliveagle/go-collectors/metadata"
	// "github.com/oliveagle/go-collectors/util"
)

const (
	ifAlias              = ".1.3.6.1.2.1.31.1.1.1.18"
	ifDescr              = ".1.3.6.1.2.1.2.2.1.2"
	ifHCInBroadcastPkts  = ".1.3.6.1.2.1.31.1.1.1.9"
	ifHCInMulticastPkts  = ".1.3.6.1.2.1.31.1.1.1.8"
	ifHCInUcastPkts      = ".1.3.6.1.2.1.31.1.1.1.7"
	ifHCOutBroadcastPkts = ".1.3.6.1.2.1.31.1.1.1.13"
	ifHCOutMulticastPkts = ".1.3.6.1.2.1.31.1.1.1.12"
	ifHCOutOctets        = ".1.3.6.1.2.1.31.1.1.1.10"
	ifHCOutUcastPkts     = ".1.3.6.1.2.1.31.1.1.1.11"
	ifHCinOctets         = ".1.3.6.1.2.1.31.1.1.1.6"
	ifInDiscards         = ".1.3.6.1.2.1.2.2.1.13"
	ifInErrors           = ".1.3.6.1.2.1.2.2.1.14"
	ifName               = ".1.3.6.1.2.1.31.1.1.1.1"
	ifOutDiscards        = ".1.3.6.1.2.1.2.2.1.19"
	ifOutErrors          = ".1.3.6.1.2.1.2.2.1.20"
)

// SNMPIfaces registers a SNMP Interfaces collector for the given community and host.
func SNMPIfaces(community, host string) {
	collectors = append(collectors, &IntervalCollector{
		F: func() (datapoint.MultiDataPoint, error) {
			return c_snmp_ifaces(community, host)
		},
		Interval: time.Second * 30,
		name:     fmt.Sprintf("snmp-ifaces-%s", host),
	})
}

func switch_bond(metric, iname string) string {
	if strings.Contains(iname, "port-channel") {
		return "os.net.bond" + strings.TrimPrefix(metric, "os.net")
	}
	return metric
}

func c_snmp_ifaces(community, host string) (datapoint.MultiDataPoint, error) {
	n, err := snmp_subtree(host, community, ifName)
	if err != nil || len(n) == 0 {
		n, err = snmp_subtree(host, community, ifDescr)
		if err != nil {
			return nil, err
		}
	}
	a, err := snmp_subtree(host, community, ifAlias)
	if err != nil {
		return nil, err
	}
	names := make(map[interface{}]string, len(n))
	aliases := make(map[interface{}]string, len(a))
	for k, v := range n {
		names[k] = fmt.Sprintf("%s", v)
	}
	for k, v := range a {
		// In case clean would come up empty, prevent the point from being removed
		// by setting our own empty case.
		aliases[k], _ = datapoint.Clean(fmt.Sprintf("%s", v))
		if aliases[k] == "" {
			aliases[k] = "NA"
		}
	}
	var md datapoint.MultiDataPoint
	add := func(oid, metric, dir string) error {
		m, err := snmp_subtree(host, community, oid)
		if err != nil {
			return err
		}
		for k, v := range m {
			tags := datapoint.TagSet{
				"host":      host,
				"direction": dir,
				"iface":     fmt.Sprintf("%d", k),
				"iname":     names[k],
			}
			Add(&md, switch_bond(metric, names[k]), v, tags, metadata.Unknown, metadata.None, "")
			metadata.AddMeta("", tags, "alias", aliases[k], false)
		}
		return nil
	}
	oids := []snmpAdd{
		{ifHCInBroadcastPkts, osNetBroadcast, "in"},
		{ifHCInMulticastPkts, osNetMulticast, "in"},
		{ifHCInUcastPkts, osNetUnicast, "in"},
		{ifHCOutBroadcastPkts, osNetBroadcast, "out"},
		{ifHCOutMulticastPkts, osNetMulticast, "out"},
		{ifHCOutOctets, osNetBytes, "out"},
		{ifHCOutUcastPkts, osNetUnicast, "out"},
		{ifHCinOctets, osNetBytes, "in"},
		{ifInDiscards, osNetDropped, "in"},
		{ifInErrors, osNetErrors, "in"},
		{ifOutDiscards, osNetDropped, "out"},
		{ifOutErrors, osNetErrors, "out"},
	}
	for _, o := range oids {
		if err := add(o.oid, o.metric, o.dir); err != nil {
			return nil, err
		}
	}
	return md, nil
}

type snmpAdd struct {
	oid    string
	metric string
	dir    string
}
