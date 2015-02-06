package collectors

import (
	"strconv"
	"strings"

	"github.com/oliveagle/go-collectors/datapoint"
	"github.com/oliveagle/go-collectors/metadata"
	"github.com/oliveagle/go-collectors/util"
)

func init() {
	collectors = append(collectors, &IntervalCollector{F: c_dfstat_darwin})
}

func c_dfstat_darwin() (datapoint.MultiDataPoint, error) {
	var md datapoint.MultiDataPoint
	util.ReadCommand(func(line string) error {
		fields := strings.Fields(line)
		if line == "" || len(fields) < 9 || !IsDigit(fields[2]) {
			return nil
		}
		mount := fields[8]
		if strings.HasPrefix(mount, "/Volumes/Time Machine Backups") {
			return nil
		}
		f1, _ := strconv.Atoi(fields[1])
		f2, _ := strconv.Atoi(fields[2])
		f3, _ := strconv.Atoi(fields[3])
		f5, _ := strconv.Atoi(fields[5])
		f6, _ := strconv.Atoi(fields[6])
		tags := datapoint.TagSet{"mount": mount}
		Add(&md, "darwin.disk.fs.total", f1, tags, metadata.Unknown, metadata.None, "")
		Add(&md, "darwin.disk.fs.used", f2, tags, metadata.Unknown, metadata.None, "")
		Add(&md, "darwin.disk.fs.free", f3, tags, metadata.Unknown, metadata.None, "")
		Add(&md, "darwin.disk.fs.inodes.total", f5+f6, tags, metadata.Unknown, metadata.None, "")
		Add(&md, "darwin.disk.fs.inodes.used", f5, tags, metadata.Unknown, metadata.None, "")
		Add(&md, "darwin.disk.fs.inodes.free", f6, tags, metadata.Unknown, metadata.None, "")
		return nil
	}, "df", "-lki")
	return md, nil
}
