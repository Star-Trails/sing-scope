package domain


// TimeSeriesPoint represents a point in time-series rate history.
type TimeSeriesPoint struct {
	Timestamp    int64   `json:"timestamp"`    // Unix ms
	UploadRate   float64 `json:"uploadRate"`   // bytes/sec
	DownloadRate float64 `json:"downloadRate"` // bytes/sec
	ActiveFlows  int     `json:"activeFlows"`
}

// NamedAggregate represents an aggregated entity with traffic metrics.
type NamedAggregate struct {
	Key             string  `json:"key"`
	Name            string  `json:"name"`
	Category        string  `json:"category,omitempty"`
	ConnectionCount int     `json:"connectionCount"`
	ActiveCount     int     `json:"activeCount"`
	UploadTotal     int64   `json:"uploadTotal"`
	DownloadTotal   int64   `json:"downloadTotal"`
	TotalBytes      int64   `json:"totalBytes"`
	UploadRate      float64 `json:"uploadRate"`
	DownloadRate    float64 `json:"downloadRate"`
	TotalRate       float64 `json:"totalRate"`
	LastActiveAt    int64   `json:"lastActiveAt"` // Unix ms
}

// ProcessAggregate represents process-specific traffic breakdown.
type ProcessAggregate struct {
	ProcessName     string           `json:"processName"`
	ProcessPath     string           `json:"processPath"`
	ProcessID       uint32           `json:"processId"`
	ConnectionCount int              `json:"connectionCount"`
	ActiveCount     int              `json:"activeCount"`
	UploadTotal     int64            `json:"uploadTotal"`
	DownloadTotal   int64            `json:"downloadTotal"`
	TotalBytes      int64            `json:"totalBytes"`
	UploadRate      float64          `json:"uploadRate"`
	DownloadRate    float64          `json:"downloadRate"`
	TopDomains      []NamedAggregate `json:"topDomains"`
	TopDestinations []NamedAggregate `json:"topDestinations"`
}

// OverviewSummary holds real-time summary cards for the Overview page.
type OverviewSummary struct {
	UploadRate       float64           `json:"uploadRate"`       // bytes/s
	DownloadRate     float64           `json:"downloadRate"`     // bytes/s
	SessionUpload    int64             `json:"sessionUpload"`    // bytes
	SessionDownload  int64             `json:"sessionDownload"`  // bytes
	ActiveTUNFlows   int               `json:"activeTunFlows"`
	ActiveTotalFlows int               `json:"activeTotalFlows"`
	TCPCount         int               `json:"tcpCount"`
	UDPCount         int               `json:"udpCount"`
	TopProcess       *NamedAggregate   `json:"topProcess,omitempty"`
	TopDomain        *NamedAggregate   `json:"topDomain,omitempty"`
	TopDestination   *NamedAggregate   `json:"topDestination,omitempty"`
	TopOutbound      *NamedAggregate   `json:"topOutbound,omitempty"`
	TimeSeries       []TimeSeriesPoint `json:"timeSeries"`
}
