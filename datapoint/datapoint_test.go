package datapoint

import (
	"encoding/json"
	"testing"
	"time"
)

func Test_MarshalDataPoint(t *testing.T) {
	d := DataPoint{
		Metric:    "metric1",
		Timestamp: time.Now().UnixNano(),
		Value:     1,
	}

	dump, _ := d.MarshalJSON()
	t.Logf("%s", dump)
	// t.Error("----")
}

func Test_UnmarshalDataPoint(t *testing.T) {
	dump := `{"metric":"metric1","timestamp":1425887018598908585,"value":1,"tags":null}`

	var v DataPoint
	json.Unmarshal([]byte(dump), &v)
	t.Log(v)
	// t.Error("---")
}

func Test_MarshalDataPoints(t *testing.T) {
	d1 := DataPoint{
		Metric:    "metric1",
		Timestamp: time.Now().UnixNano(),
		Value:     1,
	}
	d2 := DataPoint{
		Metric:    "metric2",
		Timestamp: time.Now().UnixNano(),
		Value:     2,
	}
	md := MultiDataPoint{&d1, &d2}
	t.Log(md)

	dump, err := json.Marshal(md)
	t.Log(err)
	t.Logf("%s", dump)
	// t.Error("---")

}
