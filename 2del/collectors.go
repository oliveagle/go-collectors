package collectors

// import (
// 	"bufio"
// 	"os"
// 	"sync"
// 	"time"

// 	"github.com/oliveagle/go-collectors/datapoint"
// 	"github.com/oliveagle/go-collectors/metadata"
// 	"github.com/oliveagle/go-collectors/util"
// )

// type Collector interface {
// 	Run(chan<- *datapoint.DataPoint)
// 	Name() string
// 	Init()
// }

// const (
// 	osCPU          = "os.cpu"
// 	osDiskFree     = "os.disk.fs.space_free"
// 	osDiskPctFree  = "os.disk.fs.percent_free"
// 	osDiskTotal    = "os.disk.fs.space_total"
// 	osDiskUsed     = "os.disk.fs.space_used"
// 	osMemFree      = "os.mem.free"
// 	osMemPctFree   = "os.mem.percent_free"
// 	osMemTotal     = "os.mem.total"
// 	osMemUsed      = "os.mem.used"
// 	osNetBroadcast = "os.net.packets_broadcast"
// 	osNetBytes     = "os.net.bytes"
// 	osNetDropped   = "os.net.dropped"
// 	osNetErrors    = "os.net.errs"
// 	osNetMulticast = "os.net.packets_multicast"
// 	osNetPackets   = "os.net.packets"
// 	osNetUnicast   = "os.net.packets_unicast"
// 	osSystemUptime = "os.system.uptime"
// )

// const (
// 	osDiskFreeDesc     = "The space_free property indicates in bytes how much free space is available on the disk."
// 	osDiskPctFreeDesc  = "The percent_free property indicates what percentage of the disk is available."
// 	osDiskTotalDesc    = "The space_total property indicates in bytes how much total space is on the disk."
// 	osDiskUsedDesc     = "The space_used property indicates in bytes how much space is used on the disk."
// 	osMemFreeDesc      = "Number, in bytes, of physical memory currently unused and available."
// 	osMemPctFreeDesc   = "The percent of free memory. In Linux free memory includes memory used by buffers and cache."
// 	osMemUsedDesc      = "The amount of used memory. In Linux this excludes memory used by buffers and cache."
// 	osNetBytesDesc     = "The rate at which bytes are sent or received over each network adapter."
// 	osNetDroppedDesc   = "The number of packets that were chosen to be discarded even though no errors had been detected to prevent transmission."
// 	osNetErrorsDesc    = "The number of packets that could not be transmitted because of errors."
// 	osNetPacketsDesc   = "The rate at which packets are sent or received on the network interface."
// 	osSystemUptimeDesc = "Seconds since last reboot."
// )

// var (
// 	// DefaultFreq is the duration between collection intervals if none is
// 	// specified.
// 	DefaultFreq = time.Second * 15

// 	timestamp = time.Now().Unix()
// 	tlock     sync.Mutex
// 	AddTags   datapoint.TagSet
// )

// // AddTS is the same as Add but lets you specify the timestamp
// func AddTS(md *datapoint.MultiDataPoint, name string, ts int64, value interface{}, t datapoint.TagSet, rate metadata.RateType, unit metadata.Unit, desc string) {
// 	tags := t.Copy()
// 	if rate != metadata.Unknown {
// 		metadata.AddMeta(name, nil, "rate", rate, false)
// 	}
// 	if unit != metadata.None {
// 		metadata.AddMeta(name, nil, "unit", unit, false)
// 	}
// 	if desc != "" {
// 		metadata.AddMeta(name, tags, "desc", desc, false)
// 	}
// 	if host, present := tags["host"]; !present {
// 		tags["host"] = util.Hostname
// 	} else if host == "" {
// 		delete(tags, "host")
// 	}
// 	tags = AddTags.Copy().Merge(tags)
// 	d := datapoint.DataPoint{
// 		Metric:    name,
// 		Timestamp: ts,
// 		Value:     value,
// 		Tags:      tags,
// 	}
// 	*md = append(*md, &d)
// }

// // Add appends a new data point with given metric name, value, and tags. Tags
// // may be nil. If tags is nil or does not contain a host key, it will be
// // automatically added. If the value of the host key is the empty string, it
// // will be removed (use this to prevent the normal auto-adding of the host tag).
// func Add(md *datapoint.MultiDataPoint, name string, value interface{}, t datapoint.TagSet, rate metadata.RateType, unit metadata.Unit, desc string) {
// 	AddTS(md, name, util.Now(), value, t, rate, unit, desc)
// }
