//go:build cgo

package ffi

/*
#cgo CFLAGS: -I${SRCDIR}/../../rust/traffic-core/include
#cgo LDFLAGS: ${SRCDIR}/../../rust/traffic-core/target/release/libtraffic_core.a -lm -ldl -lpthread
#include <stdlib.h>

const char* traffic_core_version();
void traffic_core_free_string(char* ptr);
char* traffic_core_analyze_batch(const char* input_json);
char* traffic_core_downsample_timeseries(const char* input_json, size_t target_buckets);
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"sync"
	"unsafe"
)

var (
	isAvailable = true
	initOnce    sync.Once
	coreVersion string
)

func initFFI() {
	initOnce.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				isAvailable = false
			}
		}()
		cVer := C.traffic_core_version()
		if cVer != nil {
			coreVersion = C.GoString(cVer)
		} else {
			coreVersion = "cgo-native"
		}
	})
}

// IsNativeCoreAvailable returns true if the Rust CGO core is loaded and functional.
func IsNativeCoreAvailable() bool {
	initFFI()
	return isAvailable
}

// GetCoreVersion returns the Rust analytics core version.
func GetCoreVersion() string {
	initFFI()
	return coreVersion
}

// NativeAnalyzeBatch invokes the Rust analytics core via coarse-grained CGO.
func NativeAnalyzeBatch(req any) ([]byte, error) {
	initFFI()
	if !isAvailable {
		return nil, fmt.Errorf("Rust native FFI is not available")
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch request: %w", err)
	}

	cInput := C.CString(string(reqBytes))
	defer C.free(unsafe.Pointer(cInput))

	cOutput := C.traffic_core_analyze_batch(cInput)
	if cOutput == nil {
		return nil, fmt.Errorf("Rust core returned null result")
	}
	defer C.traffic_core_free_string(cOutput)

	outBytes := []byte(C.GoString(cOutput))
	return outBytes, nil
}

// NativeDownsampleTimeSeries invokes the Rust downsampler via CGO.
func NativeDownsampleTimeSeries(points any, targetBuckets int) ([]byte, error) {
	initFFI()
	if !isAvailable {
		return nil, fmt.Errorf("Rust native FFI is not available")
	}

	reqBytes, err := json.Marshal(points)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal timeseries: %w", err)
	}

	cInput := C.CString(string(reqBytes))
	defer C.free(unsafe.Pointer(cInput))

	cOutput := C.traffic_core_downsample_timeseries(cInput, C.size_t(targetBuckets))
	if cOutput == nil {
		return nil, fmt.Errorf("Rust core returned null result")
	}
	defer C.traffic_core_free_string(cOutput)

	outBytes := []byte(C.GoString(cOutput))
	return outBytes, nil
}
