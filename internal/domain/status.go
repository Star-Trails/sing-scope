package domain

import (
	"time"
)

// SystemStatus represents the overall sing-box service runtime status.
type SystemStatus struct {
	Memory          uint64  `json:"memory"`          // heap memory in bytes
	Goroutines      int32   `json:"goroutines"`      // current goroutines count
	ConnectionsIn   int32   `json:"connectionsIn"`   // active inbound connections
	ConnectionsOut  int32   `json:"connectionsOut"`  // active outbound connections
	TrafficAvailable bool    `json:"trafficAvailable"`
	Uplink          int64   `json:"uplink"`          // current upload speed bytes/sec
	Downlink        int64   `json:"downlink"`        // current download speed bytes/sec
	UplinkTotal     int64   `json:"uplinkTotal"`     // lifetime total upload bytes
	DownlinkTotal   int64   `json:"downlinkTotal"`   // lifetime total download bytes
	Timestamp       time.Time `json:"timestamp"`
}

// ServiceLifecycleState represents the sing-box daemon lifecycle status.
type ServiceLifecycleState string

const (
	ServiceStateIdle     ServiceLifecycleState = "IDLE"
	ServiceStateStarting ServiceLifecycleState = "STARTING"
	ServiceStateStarted  ServiceLifecycleState = "STARTED"
	ServiceStateStopping ServiceLifecycleState = "STOPPING"
	ServiceStateFatal    ServiceLifecycleState = "FATAL"
)

// ServiceStatusInfo represents the sing-box service health & lifecycle.
type ServiceStatusInfo struct {
	State        ServiceLifecycleState `json:"state"`
	ErrorMessage string                `json:"errorMessage,omitempty"`
	Timestamp    time.Time             `json:"timestamp"`
}
