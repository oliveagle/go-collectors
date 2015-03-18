package pdh_cgo

import (
	"testing"
)

func Test_Main(t *testing.T) {

	query, err := PdhOpenQuery()
	if err != nil {
		t.Log(err)
		return
	}

	counter := PdhAddCounter(query, "\\System\\Processes")

	PdhCollectQueryData(query) // No error checking the first time

	err = PdhCollectQueryData(query)
	if err != nil {
		t.Log(err)
	}

	v, err := PdhGetDoubleCounterValue(counter)
	t.Logf("system processes count: %.0f, err: %v\n", v, err)

	if v <= 0 {
		t.Error("xxxxxx")
	}

	// t.Error("---")
}
