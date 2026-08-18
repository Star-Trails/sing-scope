package singboxapi

import (
	"fmt"
)

const (
	// TargetAPIVersion is the sing-box StartedService API version this client is built against.
	TargetAPIVersion int32 = 4

	// MinSupportedAPIVersion is the lowest API version with compatible streaming semantics.
	MinSupportedAPIVersion int32 = 4

	// MaxSupportedAPIVersion is the highest API version tested.
	MaxSupportedAPIVersion int32 = 4
)

// CompatibilityResult holds the assessment of server API version.
type CompatibilityResult struct {
	Compatible     bool   `json:"compatible"`
	ServerVersion  string `json:"serverVersion"`
	APIVersion     int32  `json:"apiVersion"`
	Message        string `json:"message"`
	Degraded       bool   `json:"degraded"`
}

// CheckCompatibility evaluates whether the remote sing-box instance version is compatible.
func CheckCompatibility(serverVersion string, apiVersion int32) CompatibilityResult {
	if apiVersion == TargetAPIVersion {
		return CompatibilityResult{
			Compatible:    true,
			ServerVersion: serverVersion,
			APIVersion:    apiVersion,
			Message:       fmt.Sprintf("sing-box %s (API v%d) fully compatible", serverVersion, apiVersion),
			Degraded:      false,
		}
	}

	if apiVersion < MinSupportedAPIVersion {
		return CompatibilityResult{
			Compatible:    false,
			ServerVersion: serverVersion,
			APIVersion:    apiVersion,
			Message:       fmt.Sprintf("sing-box API v%d is incompatible (minimum required: v%d, server version: %s)", apiVersion, MinSupportedAPIVersion, serverVersion),
			Degraded:      false,
		}
	}

	// Newer API version: attempt best effort
	return CompatibilityResult{
		Compatible:    true,
		ServerVersion: serverVersion,
		APIVersion:    apiVersion,
		Message:       fmt.Sprintf("sing-box API v%d is newer than tested v%d; running in forward-compatibility mode", apiVersion, TargetAPIVersion),
		Degraded:      true,
	}
}
