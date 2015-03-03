package collectors

import (
	"fmt"

	"github.com/oliveagle/go-collectors/datapoint"
	"github.com/oliveagle/go-collectors/metadata"
	"github.com/oliveagle/go-collectors/pdh"
	"github.com/oliveagle/go-collectors/util"
)

func init() {
	collectors = append(collectors, &IntervalCollector{F: c_pdh_windows})
}

type PdhCollector struct {
	handle   uintptr
	counters map[string]uintptr
}

func NewPdhCollector() *PdhCollector {
	var handle uintptr
	pdh.PdhOpenQuery(0, 0, &handle)

	return &PdhCollector{
		Handle: handle,
	}
}

func (p *PdhCollector) GetHandle() {
	return p.handle
}

func (p *PdhCollector) Close() {
	pdh.PdhCloseQuery(p.handle)
}

func (p *PdhCollector) AddEnglishCounter(query string) {
	pdh.PdhAddEnglishCounter(p.handle, query, 0, &p.counters[query])
}

func c_pdh_windows() (datapoint.MultiDataPoint, error) {
	var md datapoint.MultiDataPoint

	var handle uintptr
	// cHandles := make([]uintptr, 3)
	pdh.PdhOpenQuery(0, 0, &handle)

	cHandles := make(map[string]uintptr)
	// cHandles[]

	pdh.PdhAddEnglishCounter(handle, "\\System\\Processes", 0, &cHandles["\\System\\Processes"])
	pdh.PdhAddEnglishCounter(handle, "\\LogicalDisk(C:)\\% Free Space", 0, &cHandles["\\LogicalDisk(C:)\\% Free Space"])
	pdh.PdhAddEnglishCounter(handle, "\\Memory\\Available MBytes", 0, &cHandles["\\Memory\\Available MBytes"])

	pdh.PdhCollectQueryData(handle)

	var perf pdh.PDH_FMT_COUNTERVALUE_DOUBLE
	for i := 0; i < 3; i++ {
		ret = pdh.PdhGetFormattedCounterValueDouble(cHandles[i], 0, &perf)
		fmt.Printf("ret: %v, perf: %v", ret, perf) // return code will be ERROR_SUCCESS
		// pretty.Println(perf)

		if perf.DoubleValue <= 0 {
			fmt.Println("perf.DoubleValue <= 0")
		}
	}
}
