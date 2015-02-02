package datapoint

import (
	"encoding/json"
)

// DataPoint is a data point for the /api/put route:
// http://opentsdb.net/docs/build/html/api_http/put.html#example-single-data-point-put.
type DataPoint struct {
	Metric    string      `json:"metric"`
	Timestamp int64       `json:"timestamp"`
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
		Timestamp int64       `json:"timestamp"`
		Value     interface{} `json:"value"`
		Tags      TagSet      `json:"tags"`
	}{
		d.Metric,
		d.Timestamp,
		d.Value,
		d.Tags,
	})
}

// MultiDataPoint holds multiple DataPoints:
// http://opentsdb.net/docs/build/html/api_http/put.html#example-multiple-data-point-put.
type MultiDataPoint []*DataPoint
