package collectors

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/oliveagle/go-collectors/datapoint"
	"github.com/oliveagle/go-collectors/metadata"
	"github.com/oliveagle/go-collectors/util"
)

var diskLinuxFields = []struct {
	key  string
	rate metadata.RateType
	unit metadata.Unit
	desc string
}{
	{"read_requests", metadata.Counter, metadata.Count, "Total number of reads completed successfully."},
	{"read_merged", metadata.Counter, metadata.Count, "Adjacent read requests merged in a single req."},
	{"read_sectors", metadata.Counter, metadata.Count, "Total number of sectors read successfully."},
	{"msec_read", metadata.Counter, metadata.MilliSecond, "Total number of ms spent by all reads."},
	{"write_requests", metadata.Counter, metadata.Count, "Total number of writes completed successfully."},
	{"write_merged", metadata.Counter, metadata.Count, " Adjacent write requests merged in a single req."},
	{"write_sectors", metadata.Counter, metadata.Count, "Total number of sectors written successfully."},
	{"msec_write", metadata.Counter, metadata.MilliSecond, "Total number of ms spent by all writes."},
	{"ios_in_progress", metadata.Gauge, metadata.Operation, "Number of actual I/O requests currently in flight."},
	{"msec_total", metadata.Counter, metadata.MilliSecond, "Amount of time during which ios_in_progress >= 1."},
	{"msec_weighted_total", metadata.Gauge, metadata.MilliSecond, "Measure of recent I/O completion time and backlog."},
}

var diskLinuxFieldsPart = []struct {
	key  string
	rate metadata.RateType
	unit metadata.Unit
}{
	{"read_issued", metadata.Counter, metadata.Count},
	{"read_sectors", metadata.Counter, metadata.Count},
	{"write_issued", metadata.Counter, metadata.Count},
	{"write_sectors", metadata.Counter, metadata.Count},
}

func removable(major, minor string) bool {
	//We don't return an error, because removable may not exist for partitions of a removable device
	//So this is really "best effort" and we will have to see how it works in practice.
	b, err := ioutil.ReadFile("/sys/dev/block/" + major + ":" + minor + "/removable")
	if err != nil {
		return false
	}
	return strings.Trim(string(b), "\n") == "1"
}

var sdiskRE = regexp.MustCompile(`/dev/(sd[a-z])[0-9]?`)

func removable_fs(name string) bool {
	s := sdiskRE.FindStringSubmatch(name)
	if len(s) > 1 {
		b, err := ioutil.ReadFile("/sys/block/" + s[1] + "/removable")
		if err != nil {
			return false
		}
		return strings.Trim(string(b), "\n") == "1"
	}
	return false
}

func c_iostat_linux() (datapoint.MultiDataPoint, error) {
	var md datapoint.MultiDataPoint
	var removables []string
	err := readLine("/proc/diskstats", func(s string) error {
		values := strings.Fields(s)
		if len(values) < 4 {
			return nil
		} else if values[3] == "0" {
			// Skip disks that haven't done a single read.
			return nil
		}
		metric := "linux.disk.part."
		i0, _ := strconv.Atoi(values[0])
		i1, _ := strconv.Atoi(values[1])
		var block_size int64
		device := values[2]
		ts := datapoint.TagSet{"dev": device}
		if i1%16 == 0 && i0 > 1 {
			metric = "linux.disk."
			if b, err := ioutil.ReadFile("/sys/block/" + device + "/queue/hw_sector_size"); err == nil {
				block_size, _ = strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
			}
		}
		if removable(values[0], values[1]) {
			removables = append(removables, device)
		}
		for _, r := range removables {
			if strings.HasPrefix(device, r) {
				metric += "rem."
			}
		}
		if len(values) == 14 {
			var read_sectors, msec_read, write_sectors, msec_write float64
			for i, v := range values[3:] {
				switch diskLinuxFields[i].key {
				case "read_sectors":
					read_sectors, _ = strconv.ParseFloat(v, 64)
				case "msec_read":
					msec_read, _ = strconv.ParseFloat(v, 64)
				case "write_sectors":
					write_sectors, _ = strconv.ParseFloat(v, 64)
				case "msec_write":
					msec_write, _ = strconv.ParseFloat(v, 64)
				}
				Add(&md, metric+diskLinuxFields[i].key, v, ts, diskLinuxFields[i].rate, diskLinuxFields[i].unit, diskLinuxFields[i].desc)
			}
			if read_sectors != 0 && msec_read != 0 {
				Add(&md, metric+"time_per_read", read_sectors/msec_read, ts, metadata.Rate, metadata.MilliSecond, "")
			}
			if write_sectors != 0 && msec_write != 0 {
				Add(&md, metric+"time_per_write", write_sectors/msec_write, ts, metadata.Rate, metadata.MilliSecond, "")
			}
			if block_size != 0 {
				Add(&md, metric+"bytes", int64(write_sectors)*block_size, datapoint.TagSet{"type": "write"}.Merge(ts), metadata.Counter, metadata.Bytes, "Total number of bytes written to disk.")
				Add(&md, metric+"bytes", int64(read_sectors)*block_size, datapoint.TagSet{"type": "read"}.Merge(ts), metadata.Counter, metadata.Bytes, "Total number of bytes read to disk.")
				Add(&md, metric+"block_size", block_size, ts, metadata.Gauge, metadata.Bytes, "Sector size of the block device.")
			}
		} else if len(values) == 7 {
			for i, v := range values[3:] {
				Add(&md, metric+diskLinuxFieldsPart[i].key, v, ts, diskLinuxFieldsPart[i].rate, diskLinuxFieldsPart[i].unit, "")
			}
		} else {
			return fmt.Errorf("cannot parse")
		}
		return nil
	})
	return md, err
}

func c_dfstat_blocks_linux() (datapoint.MultiDataPoint, error) {
	var md datapoint.MultiDataPoint
	err := util.ReadCommand(func(line string) error {
		fields := strings.Fields(line)
		// TODO: support mount points with spaces in them. They mess up the field order
		// currently due to df's columnar output.
		if len(fields) != 6 || !IsDigit(fields[2]) {
			return nil
		}
		fs := fields[0]
		mount := fields[5]
		tags := datapoint.TagSet{"mount": mount}
		os_tags := datapoint.TagSet{"disk": mount}
		metric := "linux.disk.fs."
		ometric := "os.disk.fs."
		if removable_fs(fs) {
			metric += "rem."
			ometric += "rem."
		}
		Add(&md, metric+"space_total", fields[1], tags, metadata.Gauge, metadata.Bytes, osDiskTotalDesc)
		Add(&md, metric+"space_used", fields[2], tags, metadata.Gauge, metadata.Bytes, osDiskUsedDesc)
		Add(&md, metric+"space_free", fields[3], tags, metadata.Gauge, metadata.Bytes, osDiskFreeDesc)
		Add(&md, ometric+"space_total", fields[1], os_tags, metadata.Gauge, metadata.Bytes, osDiskTotalDesc)
		Add(&md, ometric+"space_used", fields[2], os_tags, metadata.Gauge, metadata.Bytes, osDiskUsedDesc)
		Add(&md, ometric+"space_free", fields[3], os_tags, metadata.Gauge, metadata.Bytes, osDiskFreeDesc)
		st, _ := strconv.ParseFloat(fields[1], 64)
		sf, _ := strconv.ParseFloat(fields[3], 64)
		if st != 0 {
			Add(&md, osDiskPctFree, sf/st*100, os_tags, metadata.Gauge, metadata.Pct, osDiskPctFreeDesc)
		}
		return nil
	}, "df", "-lP", "--block-size", "1")
	return md, err
}

func c_dfstat_inodes_linux() (datapoint.MultiDataPoint, error) {
	var md datapoint.MultiDataPoint
	err := util.ReadCommand(func(line string) error {
		fields := strings.Fields(line)
		if len(fields) != 6 || !IsDigit(fields[2]) {
			return nil
		}
		mount := fields[5]
		fs := fields[0]
		tags := datapoint.TagSet{"mount": mount}
		metric := "linux.disk.fs."
		if removable_fs(fs) {
			metric += "rem."
		}
		Add(&md, metric+"inodes_total", fields[1], tags, metadata.Gauge, metadata.Count, "")
		Add(&md, metric+"inodes_used", fields[2], tags, metadata.Gauge, metadata.Count, "")
		Add(&md, metric+"inodes_free", fields[3], tags, metadata.Gauge, metadata.Count, "")
		return nil
	}, "df", "-liP")
	return md, err
}
