package datapoint

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"
)

var bigMaxInt64 = big.NewInt(math.MaxInt64)

// DataPoint is a data point for the /api/put route:
// http://opentsdb.net/docs/build/html/api_http/put.html#example-single-data-point-put.
type DataPoint struct {
	Metric string `json:"metric"`
	// Timestamp int64       `json:"timestamp"`
	Timestamp time.Time   `json:"timestamp"`
	Value     interface{} `json:"value"`
	Tags      TagSet      `json:"tags"`
}

// MarshalJSON verifies d is valid and converts it to JSON.
func (d *DataPoint) MarshalJSON() ([]byte, error) {
	if err := d.clean(); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Metric    string      `json:"metric"`
		Timestamp time.Time   `json:"timestamp"`
		Value     interface{} `json:"value"`
		Tags      TagSet      `json:"tags"`
	}{
		d.Metric,
		d.Timestamp,
		d.Value,
		d.Tags,
	})
}

func (d *DataPoint) clean() error {
	if err := d.Tags.Clean(); err != nil {
		return err
	}
	m, err := Clean(d.Metric)
	if err != nil {
		return fmt.Errorf("cleaning metric %s: %s", d.Metric, err)
	}
	if d.Metric != m {
		d.Metric = m
	}
	switch v := d.Value.(type) {
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			d.Value = i
		} else if f, err := strconv.ParseFloat(v, 64); err == nil {
			d.Value = f
		} else {
			return fmt.Errorf("Unparseable number %v", v)
		}
	case uint64:
		if v > math.MaxInt64 {
			d.Value = float64(v)
		}
	case *big.Int:
		if bigMaxInt64.Cmp(v) < 0 {
			if f, err := strconv.ParseFloat(v.String(), 64); err == nil {
				d.Value = f
			}
		}
	}
	return nil
}

// MultiDataPoint holds multiple DataPoints:
// http://opentsdb.net/docs/build/html/api_http/put.html#example-multiple-data-point-put.
type MultiDataPoint []*DataPoint
