// +build windows

package pdh_cgo

import (
	"fmt"
	"unsafe"
)

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -lpdh

#include <windows.h>
#include <stdio.h>
#include <pdh.h>
#include <pdhmsg.h>

#pragma comment(lib, "pdh.lib")


#ifndef PDH_FMT_NOCAP100
#define PDH_FMT_NOCAP100 ((DWORD) 0x00008000)
#endif

double GetDoubleValueOfPdhData(PDH_FMT_COUNTERVALUE cv) {
    return cv.doubleValue;
}
*/
import "C"

func PdhOpenQuery() (query C.HQUERY, err error) {
	status := C.PdhOpenQuery(C.LPCSTR(nil), 0, &query)
	if status != C.ERROR_SUCCESS {
		return nil, fmt.Errorf("PdhOpenQuery Failed: 0x%X\n", status)
	}
	return
}

func PdhAddCounter(query C.HQUERY, path string) (counter C.HCOUNTER) {
	cpath := C.CString(path)
	C.PdhAddCounter(query, (*C.CHAR)(cpath), 0, &counter)
	C.free(unsafe.Pointer(cpath))
	return
}

func PdhCollectQueryData(query C.HQUERY) error {
	status := C.PdhCollectQueryData(query)
	if status != C.ERROR_SUCCESS {
		if uint64(status) == uint64(C.PDH_INVALID_HANDLE) {
			return fmt.Errorf("PDH_INVALID_HANDLE")
		} else if uint64(status) == uint64(C.PDH_NO_DATA) {
			return fmt.Errorf("PDH_NO_DATA")
		}
		return fmt.Errorf("Unknown error: 0x%X", status)
	}
	return nil
}

func PdhGetDoubleCounterValue(counter C.HCOUNTER) (float64, error) {
	var ret C.DWORD
	var value C.PDH_FMT_COUNTERVALUE
	status := C.PdhGetFormattedCounterValue(counter, C.PDH_FMT_DOUBLE|C.PDH_FMT_NOCAP100|C.PDH_FMT_NOSCALE, &ret, &value)
	if status != C.ERROR_SUCCESS {
		// fmt.Printf("PdhGetFormattedCounterValue() ***Error: 0x%X\n", status)
		return 0.0, fmt.Errorf("PdhGetDoubleCounterValue() Error: 0x%X", status)
	}
	return float64(C.GetDoubleValueOfPdhData(value)), nil
}
