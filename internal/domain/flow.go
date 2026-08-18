package domain

import (
	"time"
)

// ProcessInfo contains operating-system level process metadata for a network flow.
type ProcessInfo struct {
	ProcessID    uint32   `json:"processId"`
	UserID       int32    `json:"userId"`
	UserName     string   `json:"userName"`
	ProcessPath  string   `json:"processPath"`
	ProcessName  string   `json:"processName"`
	PackageNames []string `json:"packageNames"`
}

// Flow represents a normalized, authoritative network connection flow tracked by sing-box.
type Flow struct {
	ID string `json:"id"`

	Inbound     string `json:"inbound"`     // e.g. "tun-in", "mixed-in"
	InboundType string `json:"inboundType"` // e.g. "tun", "mixed", "tproxy", "direct"

	IPVersion int    `json:"ipVersion"` // 4 or 6
	Network   string `json:"network"`   // "tcp" or "udp"

	Source      string `json:"source"`      // "ip:port"
	Destination string `json:"destination"` // "ip:port"

	Domain   string `json:"domain"`   // e.g. "example.com"
	Protocol string `json:"protocol"` // e.g. "tls", "http", "quic", "dns"
	User     string `json:"user"`

	FromOutbound string   `json:"fromOutbound"`
	Rule         string   `json:"rule"`
	Outbound     string   `json:"outbound"`
	OutboundType string   `json:"outboundType"`
	ChainList    []string `json:"chainList"`

	Process *ProcessInfo `json:"process,omitempty"`

	CreatedAt time.Time  `json:"createdAt"`
	ClosedAt  *time.Time `json:"closedAt,omitempty"`

	// Counters and Rates
	UploadTotal   int64 `json:"uploadTotal"`
	DownloadTotal int64 `json:"downloadTotal"`

	UploadDelta   int64 `json:"uploadDelta"`
	DownloadDelta int64 `json:"downloadDelta"`

	UploadRate   float64 `json:"uploadRate"`   // bytes/second
	DownloadRate float64 `json:"downloadRate"` // bytes/second

	LastActiveAt time.Time `json:"lastActiveAt"`
	IsActive     bool      `json:"isActive"`
}

// FlowEventType defines the type of event occurring on a flow.
type FlowEventType string

const (
	FlowEventNew    FlowEventType = "NEW"
	FlowEventUpdate FlowEventType = "UPDATE"
	FlowEventClosed FlowEventType = "CLOSED"
	FlowEventReset  FlowEventType = "RESET"
)

// FlowEvent represents a delta or state change on a flow.
type FlowEvent struct {
	Type          FlowEventType `json:"type"`
	ID            string        `json:"id"`
	Flow          *Flow         `json:"flow,omitempty"`
	UplinkDelta   int64         `json:"uplinkDelta"`
	DownlinkDelta int64         `json:"downlinkDelta"`
	ClosedAt      *time.Time    `json:"closedAt,omitempty"`
	Timestamp     time.Time     `json:"timestamp"`
}
