package collectors

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oliveagle/go-collectors/datapoint"
	"github.com/oliveagle/go-collectors/metadata"
	"github.com/oliveagle/go-collectors/util"
)

func init() {
	collectors = append(collectors, &IntervalCollector{F: c_iostat_darwin})
}

func c_iostat_darwin() (datapoint.MultiDataPoint, error) {
	var categories []string
	var md datapoint.MultiDataPoint
	ln := 0
	i := 0
	util.ReadCommand(func(line string) error {
		ln++
		if ln == 1 {
			categories = strings.Fields(line)
		}
		if ln < 4 {
			return nil
		}
		values := strings.Fields(line)
		for _, cat := range categories {
			if i+3 > len(values) {
				break
			} else if strings.HasPrefix(cat, "disk") {
				Add(&md, "darwin.disk.kilobytes_transfer", values[i], datapoint.TagSet{"disk": cat}, metadata.Unknown, metadata.None, "")
				i++
				Add(&md, "darwin.disk.transactions", values[i], datapoint.TagSet{"disk": cat}, metadata.Unknown, metadata.None, "")
				i++
				Add(&md, "darwin.disk.megabytes", values[i], datapoint.TagSet{"disk": cat}, metadata.Unknown, metadata.None, "")
				i++
			} else if cat == "cpu" {
				Add(&md, "darwin.cpu.user", values[i], nil, metadata.Gauge, metadata.Pct, descDarwinCPUUser)
				i++
				Add(&md, "darwin.cpu.sys", values[i], nil, metadata.Gauge, metadata.Pct, descDarwinCPUSys)
				i++
				Add(&md, "darwin.cpu.idle", values[i], nil, metadata.Gauge, metadata.Pct, descDarwinCPUIdle)
				i++
			} else if cat == "load" {
				load, _ := strconv.ParseFloat(values[i], 64)
				Add(&md, "darwin.loadavg_1_min", load, nil, metadata.Unknown, metadata.None, "")
				i++

				load, _ = strconv.ParseFloat(values[i], 64)
				Add(&md, "darwin.loadavg_5_min", load, nil, metadata.Unknown, metadata.None, "")
				i++

				load, _ = strconv.ParseFloat(values[i], 64)
				Add(&md, "darwin.loadavg_15_min", load, nil, metadata.Unknown, metadata.None, "")
				i++
			}
		}
		return nil
	}, "iostat", "-c2", "-w1")
	if ln < 4 {
		return nil, fmt.Errorf("bad return value")
	}
	return md, nil
}

const (
	descDarwinCPUUser = "Percent of time in user mode."
	descDarwinCPUSys  = "Percent of time in system mode."
	descDarwinCPUIdle = "Percent of time in idle mode."
)
