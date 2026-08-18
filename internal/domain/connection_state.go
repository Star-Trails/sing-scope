package domain

import "time"

// ConnectionState represents the gRPC client connection status.
type ConnectionState string

const (
	StateDisconnected   ConnectionState = "Disconnected"
	StateConnecting     ConnectionState = "Connecting"
	StateConnected      ConnectionState = "Connected"
	StateReconnecting   ConnectionState = "Reconnecting"
	StateAuthFailed     ConnectionState = "Authentication failed"
	StateIncompatible   ConnectionState = "API incompatible"
	StateError          ConnectionState = "Error"
)

// ServerConnectionInfo represents current connection metadata and version info.
type ServerConnectionInfo struct {
	State          ConnectionState `json:"state"`
	ServerURL      string          `json:"serverUrl"`
	SingBoxVersion string          `json:"singBoxVersion"`
	APIVersion     int32           `json:"apiVersion"`
	ErrorMessage   string          `json:"errorMessage,omitempty"`
	ConnectedAt    *time.Time      `json:"connectedAt,omitempty"`
	LastEventAt    *time.Time      `json:"lastEventAt,omitempty"`
}

// ClientConfig holds the connection parameters.
type ClientConfig struct {
	ServerURL string `json:"serverUrl"`
	Secret    string `json:"secret"` // kept securely in Go backend, not exposed unnecessarily
	IntervalMs int64 `json:"intervalMs"`
	AutoConnect bool  `json:"autoConnect"`
}
