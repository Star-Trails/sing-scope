//go:build !cgo

package ffi

import (
	"errors"
)

// IsNativeCoreAvailable returns false in pure Go builds.
func IsNativeCoreAvailable() bool {
	return false
}

// GetCoreVersion returns the engine identifier.
func GetCoreVersion() string {
	return "pure-go"
}

// NativeAnalyzeBatch returns an error indicating pure-go fallback is used.
func NativeAnalyzeBatch(req any) ([]byte, error) {
	return nil, errors.New("CGO disabled; using pure-Go analytics engine")
}

// NativeDownsampleTimeSeries returns an error indicating pure-go fallback is used.
func NativeDownsampleTimeSeries(points any, targetBuckets int) ([]byte, error) {
	return nil, errors.New("CGO disabled; using pure-Go analytics engine")
}
