// +build windows

package pdh

import (
	"time"
)

type PdhDataPoint struct {
	Query     string
	Timestamp time.Time
	Value     float64
}

type PdhCollector struct {
	handle   uintptr
	counters map[string]uintptr
}

func NewPdhCollector() *PdhCollector {
	var handle uintptr
	PdhOpenQuery(0, 0, &handle)

	return &PdhCollector{
		handle:   handle,
		counters: make(map[string]uintptr),
	}
}

func (p *PdhCollector) GetHandle() uintptr {
	return p.handle
}

func (p *PdhCollector) Close() {
	PdhCloseQuery(p.handle)
}

func (p *PdhCollector) AddEnglishCounter(query string) {
	var handle uintptr
	PdhAddEnglishCounter(p.handle, query, 0, &handle)
	p.counters[query] = handle
}

func (p *PdhCollector) CollectData() []*PdhDataPoint {
	PdhCollectQueryData(p.handle)
	data := []*PdhDataPoint{}

	var perf PDH_FMT_COUNTERVALUE_DOUBLE
	for key, chandle := range p.counters {
		PdhGetFormattedCounterValueDouble(chandle, 0, &perf)
		pd := PdhDataPoint{
			Query:     key,
			Timestamp: time.Now(),
			Value:     perf.DoubleValue,
		}
		data = append(data, &pd)
	}
	return data
}
