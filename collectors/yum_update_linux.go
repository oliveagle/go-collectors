package collectors

import (
	"strings"
	"time"

	"bosun.org/util"
	"github.com/oliveagle/go-collectors/datapoint"
	"github.com/oliveagle/go-collectors/metadata"
)

func init() {
	collectors = append(collectors, &IntervalCollector{F: yum_update_stats_linux, Interval: time.Minute * 30})
}

func yum_update_stats_linux() (datapoint.MultiDataPoint, error) {
	var md datapoint.MultiDataPoint
	regular_c := 0
	kernel_c := 0
	// This is a silly long timeout, but until we implement sigint this will
	// Prevent a currupt yum db https://github.com/bosun-monitor/scollector/issues/56
	err := util.ReadCommandTimeout(time.Minute*5, func(line string) error {
		fields := strings.Fields(line)
		if len(fields) > 1 && !strings.HasPrefix(fields[0], "Updated Packages") {
			if strings.HasPrefix(fields[0], "kern") {
				kernel_c++
			} else {
				regular_c++
			}
		}
		return nil

	}, "yum", "list", "updates", "-q")
	if err == util.ErrPath {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	Add(&md, "linux.updates.count", regular_c, datapoint.TagSet{"type": "non-kernel"}, metadata.Gauge, metadata.Count, "")
	Add(&md, "linux.updates.count", kernel_c, datapoint.TagSet{"type": "kernel"}, metadata.Gauge, metadata.Count, "")
	return md, nil
}
