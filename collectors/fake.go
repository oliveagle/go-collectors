package collectors

import (
	"strconv"
	"time"

	"github.com/oliveagle/go-collectors/datapoint"
	"github.com/oliveagle/go-collectors/metadata"
	// "github.com/oliveagle/go-collectors/util"
)

func InitFake(fake int) {
	collectors = append(collectors, &IntervalCollector{
		F: func() (datapoint.MultiDataPoint, error) {
			var md datapoint.MultiDataPoint
			for i := 0; i < fake; i++ {
				Add(&md, "test.fake", i, datapoint.TagSet{"i": strconv.Itoa(i)}, metadata.Unknown, metadata.None, "")
			}
			return md, nil
		},
		Interval: time.Second,
		name:     "fake",
	})
}
