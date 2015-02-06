package collectors

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/StackExchange/wmi"
	"github.com/oliveagle/go-collectors/datapoint"
	"github.com/oliveagle/go-collectors/metadata"
	// "github.com/oliveagle/go-collectors/util"
)

func init() {
	collectors = append(collectors, &IntervalCollector{F: c_windows_processes})
}

func WatchProcesses(procs []*WatchedProc) error {
	return fmt.Errorf("process watching not implemented on Darwin")
}

// These are silly processes but exist on my machine, will need to update KMB
var processInclusions = regexp.MustCompile("chrome|powershell|scollector|SocketServer")
var serviceInclusions = regexp.MustCompile("WinRM|MSSQLSERVER|StackServerProd|StackServerDev|LogStasher")

func c_windows_processes() (datapoint.MultiDataPoint, error) {
	var dst []Win32_PerfRawData_PerfProc_Process
	var q = wmi.CreateQuery(&dst, `WHERE Name <> '_Total'`)
	err := queryWmi(q, &dst)
	if err != nil {
		return nil, err
	}

	var svc_dst []Win32_Service
	var svc_q = wmi.CreateQuery(&svc_dst, `WHERE Name <> '_Total'`)
	err = queryWmi(svc_q, &svc_dst)
	if err != nil {
		return nil, err
	}

	var iis_dst []WorkerProcess
	iis_q := wmi.CreateQuery(&iis_dst, "")
	err = queryWmiNamespace(iis_q, &iis_dst, "root\\WebAdministration")
	if err != nil {
		// Don't return from this error since the name space might exist.
		iis_dst = nil
	}

	var numberOfLogicalProcessors uint64
	var core_dst []Win32_ComputerSystem
	var core_q = wmi.CreateQuery(&core_dst, "")
	err = queryWmi(core_q, &core_dst)
	if err != nil {
		return nil, err
	}
	for _, y := range core_dst {
		numberOfLogicalProcessors = uint64(y.NumberOfLogicalProcessors)
	}
	if numberOfLogicalProcessors == 0 {
		return nil, fmt.Errorf("invalid result: numberOfLogicalProcessors=%g", numberOfLogicalProcessors)
	}

	var md datapoint.MultiDataPoint
	for _, v := range dst {
		var name string
		service_match := false
		iis_match := false
		process_match := processInclusions.MatchString(v.Name)

		id := "0"

		if process_match {
			raw_name := strings.Split(v.Name, "#")
			name = raw_name[0]
			if len(raw_name) == 2 {
				id = raw_name[1]
			}
			// If you have a hash sign in your process name you don't deserve monitoring ;-)
			if len(raw_name) > 2 {
				continue
			}
		}

		// A Service match could "overwrite" a process match, but that is probably what we would want
		for _, svc := range svc_dst {
			if serviceInclusions.MatchString(svc.Name) {
				// It is possible the pid has gone and been reused, but I think this unlikely
				// And I'm not aware of an atomic join we could do anyways
				if svc.ProcessId == v.IDProcess {
					id = "0"
					service_match = true
					name = svc.Name
					break
				}
			}
		}

		for _, a_pool := range iis_dst {
			if a_pool.ProcessId == v.IDProcess {
				id = "0"
				iis_match = true
				name = strings.Join([]string{"iis", a_pool.AppPoolName}, "_")
				break
			}
		}

		if !(service_match || process_match || iis_match) {
			continue
		}

		//Use timestamp from WMI to fix issues with CPU metrics
		ts := TSys100NStoEpoch(v.Timestamp_Sys100NS)
		AddTS(&md, "win.proc.cpu", ts, v.PercentPrivilegedTime/NS100_Seconds/numberOfLogicalProcessors, datapoint.TagSet{"name": name, "id": id, "type": "privileged"}, metadata.Counter, metadata.Pct, descWinProcCPU_priv)
		AddTS(&md, "win.proc.cpu", ts, v.PercentUserTime/NS100_Seconds/numberOfLogicalProcessors, datapoint.TagSet{"name": name, "id": id, "type": "user"}, metadata.Counter, metadata.Pct, descWinProcCPU_user)
		AddTS(&md, "win.proc.cpu_total", ts, v.PercentProcessorTime/NS100_Seconds/numberOfLogicalProcessors, datapoint.TagSet{"name": name, "id": id}, metadata.Counter, metadata.Pct, descWinProcCPU_total)
		Add(&md, "win.proc.elapsed_time", (v.Timestamp_Object-v.ElapsedTime)/v.Frequency_Object, datapoint.TagSet{"name": name, "id": id}, metadata.Gauge, metadata.Second, descWinProcElapsed_time)
		Add(&md, "win.proc.handle_count", v.HandleCount, datapoint.TagSet{"name": name, "id": id}, metadata.Gauge, metadata.Count, descWinProcHandle_count)
		Add(&md, "win.proc.io_bytes", v.IOOtherBytesPersec, datapoint.TagSet{"name": name, "id": id, "type": "other"}, metadata.Counter, metadata.BytesPerSecond, descWinProcIo_bytes_other)
		Add(&md, "win.proc.io_bytes", v.IOReadBytesPersec, datapoint.TagSet{"name": name, "id": id, "type": "read"}, metadata.Counter, metadata.BytesPerSecond, descWinProcIo_bytes_read)
		Add(&md, "win.proc.io_bytes", v.IOWriteBytesPersec, datapoint.TagSet{"name": name, "id": id, "type": "write"}, metadata.Counter, metadata.BytesPerSecond, descWinProcIo_bytes_write)
		Add(&md, "win.proc.io_operations", v.IOOtherOperationsPersec, datapoint.TagSet{"name": name, "id": id, "type": "other"}, metadata.Counter, metadata.Operation, descWinProcIo_operations)
		Add(&md, "win.proc.io_operations", v.IOReadOperationsPersec, datapoint.TagSet{"name": name, "id": id, "type": "read"}, metadata.Counter, metadata.Operation, descWinProcIo_operations)
		Add(&md, "win.proc.io_operations", v.IOWriteOperationsPersec, datapoint.TagSet{"name": name, "id": id, "type": "write"}, metadata.Counter, metadata.Operation, descWinProcIo_operations)
		Add(&md, "win.proc.mem.page_faults", v.PageFaultsPersec, datapoint.TagSet{"name": name, "id": id}, metadata.Counter, metadata.PerSecond, descWinProcMemPage_faults)
		Add(&md, "win.proc.mem.pagefile_bytes", v.PageFileBytes, datapoint.TagSet{"name": name, "id": id}, metadata.Gauge, metadata.Bytes, descWinProcMemPagefile_bytes)
		Add(&md, "win.proc.mem.pagefile_bytes_peak", v.PageFileBytesPeak, datapoint.TagSet{"name": name, "id": id}, metadata.Gauge, metadata.Bytes, descWinProcMemPagefile_bytes_peak)
		Add(&md, "win.proc.mem.pool_nonpaged_bytes", v.PoolNonpagedBytes, datapoint.TagSet{"name": name, "id": id}, metadata.Gauge, metadata.Bytes, descWinProcMemPool_nonpaged_bytes)
		Add(&md, "win.proc.mem.pool_paged_bytes", v.PoolPagedBytes, datapoint.TagSet{"name": name, "id": id}, metadata.Gauge, metadata.Bytes, descWinProcMemPool_paged_bytes)
		Add(&md, "win.proc.mem.vm.bytes", v.VirtualBytes, datapoint.TagSet{"name": name, "id": id}, metadata.Gauge, metadata.Bytes, descWinProcMemVmBytes)
		Add(&md, "win.proc.mem.vm.bytes_peak", v.VirtualBytesPeak, datapoint.TagSet{"name": name, "id": id}, metadata.Gauge, metadata.Bytes, descWinProcMemVmBytes_peak)
		Add(&md, "win.proc.mem.working_set", v.WorkingSet, datapoint.TagSet{"name": name, "id": id}, metadata.Gauge, metadata.Bytes, descWinProcMemWorking_set)
		Add(&md, "win.proc.mem.working_set_peak", v.WorkingSetPeak, datapoint.TagSet{"name": name, "id": id}, metadata.Gauge, metadata.Bytes, descWinProcMemWorking_set_peak)
		Add(&md, "win.proc.mem.working_set_private", v.WorkingSetPrivate, datapoint.TagSet{"name": name, "id": id}, metadata.Gauge, metadata.Bytes, descWinProcMemWorking_set_private)
		Add(&md, "win.proc.priority_base", v.PriorityBase, datapoint.TagSet{"name": name, "id": id}, metadata.Gauge, metadata.None, descWinProcPriority_base)
		Add(&md, "win.proc.private_bytes", v.PrivateBytes, datapoint.TagSet{"name": name, "id": id}, metadata.Gauge, metadata.Bytes, descWinProcPrivate_bytes)
		Add(&md, "win.proc.thread_count", v.ThreadCount, datapoint.TagSet{"name": name, "id": id}, metadata.Gauge, metadata.Count, descWinProcthread_count)
	}
	return md, nil
}

// Divide CPU by 1e5 because: 1 seconds / 100 Nanoseconds = 1e7. This is the
// percent time as a decimal, so divide by two less zeros to make it the same as
// the result * 100.
const NS100_Seconds = 1e5

const (
	descWinProcCPU_priv               = "Percentage of elapsed time that this thread has spent executing code in privileged mode."
	descWinProcCPU_total              = "Percentage of elapsed time that this process's threads have spent executing code in user or privileged mode."
	descWinProcCPU_user               = "Percentage of elapsed time that this process's threads have spent executing code in user mode."
	descWinProcElapsed_time           = "Elapsed time in seconds this process has been running."
	descWinProcHandle_count           = "Total number of handles the process has open across all threads."
	descWinProcIo_bytes_other         = "Rate at which the process is issuing bytes to I/O operations that do not involve data such as control operations."
	descWinProcIo_bytes_read          = "Rate at which the process is reading bytes from I/O operations."
	descWinProcIo_bytes_write         = "Rate at which the process is writing bytes to I/O operations."
	descWinProcIo_operations          = "Rate at which the process is issuing I/O operations that are neither a read or a write request."
	descWinProcIo_operations_read     = "Rate at which the process is issuing read I/O operations."
	descWinProcIo_operations_write    = "Rate at which the process is issuing write I/O operations."
	descWinProcMemPage_faults         = "Rate of page faults by the threads executing in this process."
	descWinProcMemPagefile_bytes      = "Current number of bytes this process has used in the paging file(s)."
	descWinProcMemPagefile_bytes_peak = "Maximum number of bytes this process has used in the paging file(s)."
	descWinProcMemPool_nonpaged_bytes = "Total number of bytes for objects that cannot be written to disk when they are not being used."
	descWinProcMemPool_paged_bytes    = "Total number of bytes for objects that can be written to disk when they are not being used."
	descWinProcMemVmBytes             = "Current size, in bytes, of the virtual address space that the process is using."
	descWinProcMemVmBytes_peak        = "Maximum number of bytes of virtual address space that the process has used at any one time."
	descWinProcMemWorking_set         = "Current number of bytes in the working set of this process at any point in time."
	descWinProcMemWorking_set_peak    = "Maximum number of bytes in the working set of this process at any point in time."
	descWinProcMemWorking_set_private = "Current number of bytes in the working set that are not shared with other processes."
	descWinProcPriority_base          = "Current base priority of this process. Threads within a process can raise and lower their own base priority relative to the process base priority of the process."
	descWinProcPrivate_bytes          = "Current number of bytes this process has allocated that cannot be shared with other processes."
	descWinProcthread_count           = "Number of threads currently active in this process."
)

// Actually a CIM_StatisticalInformation.
type Win32_PerfRawData_PerfProc_Process struct {
	ElapsedTime             uint64
	Frequency_Object        uint64
	HandleCount             uint32
	IDProcess               uint32
	IOOtherBytesPersec      uint64
	IOOtherOperationsPersec uint64
	IOReadBytesPersec       uint64
	IOReadOperationsPersec  uint64
	IOWriteBytesPersec      uint64
	IOWriteOperationsPersec uint64
	Name                    string
	PageFaultsPersec        uint32
	PageFileBytes           uint64
	PageFileBytesPeak       uint64
	PercentPrivilegedTime   uint64
	PercentProcessorTime    uint64
	PercentUserTime         uint64
	PoolNonpagedBytes       uint32
	PoolPagedBytes          uint32
	PriorityBase            uint32
	PrivateBytes            uint64
	ThreadCount             uint32
	Timestamp_Object        uint64
	Timestamp_Sys100NS      uint64
	VirtualBytes            uint64
	VirtualBytesPeak        uint64
	WorkingSet              uint64
	WorkingSetPeak          uint64
	WorkingSetPrivate       uint64
}

// Actually a Win32_BaseServce.
type Win32_Service struct {
	Name      string
	ProcessId uint32
}

type WorkerProcess struct {
	AppPoolName string
	ProcessId   uint32
}

type Win32_ComputerSystem struct {
	NumberOfLogicalProcessors uint32
}
