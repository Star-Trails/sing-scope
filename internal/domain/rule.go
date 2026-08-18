package domain

// RuleInfo represents a routing rule with hit count and traffic metrics.
type RuleInfo struct {
	Type       string `json:"type"`
	Payload    string `json:"payload"`
	Proxy      string `json:"proxy"`
	HitCount   int    `json:"hitCount"`
	TotalBytes int64  `json:"totalBytes"`
	LastHitAt  int64  `json:"lastHitAt"` // Unix ms
	UUID       string `json:"uuid"`
	Index      int    `json:"index"`
}
