package store

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"sing-scope/internal/domain"
)

// StoreOptions controls history limits and retention.
type StoreOptions struct {
	MaxHistorySize   int
	MaxTimeSeriesPts int
}

// DefaultStoreOptions returns sensible defaults for high-performance memory use.
func DefaultStoreOptions() StoreOptions {
	return StoreOptions{
		MaxHistorySize:   50000,
		MaxTimeSeriesPts: 900, // 15 minutes of 1s samples
	}
}

// ConnectionStore is the authoritative state store for active flows, history, and traffic rates.
type ConnectionStore struct {
	mu sync.RWMutex

	opts StoreOptions

	// Active flows indexed by Flow ID
	activeFlows map[string]*domain.Flow

	// Historical (closed) flows in insertion order (ring buffer bounded)
	closedHistory []*domain.Flow
	historyIndex  int

	// Discovered and matched routing rules
	rules map[string]*domain.RuleInfo

	// Session-level accumulators
	sessionUploadBytes   int64
	sessionDownloadBytes int64
	currentUploadRate    float64
	currentDownloadRate  float64
	peakUploadRate       float64
	peakDownloadRate     float64

	// Time-series history (1s points)
	timeSeries []domain.TimeSeriesPoint

	// Known inbounds, outbounds, processes for filter suggestions
	inbounds   map[string]string // tag -> type
	outbounds  map[string]string // tag -> type
	processes  map[string]bool   // processName
	protocols  map[string]bool

	// Active TUN filter tag (empty = all inbounds, "tun" = all tun, or specific tag)
	selectedInboundFilter string

	lastUpdateTimestamp time.Time
}

// NewConnectionStore creates an initialized ConnectionStore.
func NewConnectionStore(opts StoreOptions) *ConnectionStore {
	if opts.MaxHistorySize <= 0 {
		opts.MaxHistorySize = 50000
	}
	if opts.MaxTimeSeriesPts <= 0 {
		opts.MaxTimeSeriesPts = 900
	}

	return &ConnectionStore{
		opts:                opts,
		activeFlows:         make(map[string]*domain.Flow),
		closedHistory:       make([]*domain.Flow, 0, 1024),
		rules:               make(map[string]*domain.RuleInfo),
		timeSeries:          make([]domain.TimeSeriesPoint, 0, opts.MaxTimeSeriesPts),
		inbounds:            make(map[string]string),
		outbounds:           make(map[string]string),
		processes:           make(map[string]bool),
		protocols:           make(map[string]bool),
		lastUpdateTimestamp: time.Now(),
	}
}

// SetInboundFilter configures active inbound filter ("" for all, "tun:all" for all TUN, or exact tag).
func (s *ConnectionStore) SetInboundFilter(filter string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selectedInboundFilter = filter
}

// GetInboundFilter returns the current inbound filter.
func (s *ConnectionStore) GetInboundFilter() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selectedInboundFilter
}

// Reset clears active connection state (used when ConnectionEvents.reset == true).
func (s *ConnectionStore) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.activeFlows = make(map[string]*domain.Flow)
	s.currentUploadRate = 0
	s.currentDownloadRate = 0
	s.lastUpdateTimestamp = time.Now()
}

// ProcessBatch applies a batch of flow events to the authoritative store.
func (s *ConnectionStore) ProcessBatch(events []domain.FlowEvent, isReset bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	var dt float64 = 1.0
	if !s.lastUpdateTimestamp.IsZero() {
		elapsed := now.Sub(s.lastUpdateTimestamp).Seconds()
		if elapsed > 0.05 {
			dt = elapsed
		}
	}
	s.lastUpdateTimestamp = now

	if isReset {
		s.activeFlows = make(map[string]*domain.Flow)
		s.currentUploadRate = 0
		s.currentDownloadRate = 0
	}

	var batchUploadDelta int64
	var batchDownloadDelta int64

	for _, event := range events {
		switch event.Type {
		case domain.FlowEventNew:
			if event.Flow != nil {
				flow := event.Flow
				flow.LastActiveAt = now

				// Track metadata catalogs
				if flow.Inbound != "" {
					s.inbounds[flow.Inbound] = flow.InboundType
				}
				if flow.Outbound != "" {
					s.outbounds[flow.Outbound] = flow.OutboundType
				}
				if flow.Process != nil && flow.Process.ProcessName != "" {
					s.processes[flow.Process.ProcessName] = true
				}
				if flow.Protocol != "" {
					s.protocols[flow.Protocol] = true
				}

				s.recordRule(flow, now)

				if flow.ClosedAt != nil {
					flow.IsActive = false
					s.appendHistory(flow)
				} else {
					flow.IsActive = true
					s.activeFlows[flow.ID] = flow
				}
			}

		case domain.FlowEventUpdate:
			flow, exists := s.activeFlows[event.ID]
			if exists {
				flow.UploadDelta = event.UplinkDelta
				flow.DownloadDelta = event.DownlinkDelta

				flow.UploadTotal += event.UplinkDelta
				flow.DownloadTotal += event.DownlinkDelta

				flow.UploadRate = float64(event.UplinkDelta) / dt
				flow.DownloadRate = float64(event.DownlinkDelta) / dt

				flow.LastActiveAt = now

				s.recordRule(flow, now)

				batchUploadDelta += event.UplinkDelta
				batchDownloadDelta += event.DownlinkDelta
			}

		case domain.FlowEventClosed:
			flow, exists := s.activeFlows[event.ID]
			if exists {
				delete(s.activeFlows, event.ID)
				closedAt := now
				if event.ClosedAt != nil {
					closedAt = *event.ClosedAt
				}
				flow.ClosedAt = &closedAt
				flow.IsActive = false
				flow.UploadRate = 0
				flow.DownloadRate = 0
				flow.UploadDelta = 0
				flow.DownloadDelta = 0
				s.recordRule(flow, now)
				s.appendHistory(flow)
			} else if event.Flow != nil {
				flow := event.Flow
				flow.IsActive = false
				s.recordRule(flow, now)
				s.appendHistory(flow)
			}
		}
	}

	// Update session totals & instantaneous rates
	s.sessionUploadBytes += batchUploadDelta
	s.sessionDownloadBytes += batchDownloadDelta

	s.currentUploadRate = float64(batchUploadDelta) / dt
	s.currentDownloadRate = float64(batchDownloadDelta) / dt

	if s.currentUploadRate > s.peakUploadRate {
		s.peakUploadRate = s.currentUploadRate
	}
	if s.currentDownloadRate > s.peakDownloadRate {
		s.peakDownloadRate = s.currentDownloadRate
	}

	// Record time series point
	tsPoint := domain.TimeSeriesPoint{
		Timestamp:    now.UnixMilli(),
		UploadRate:   s.currentUploadRate,
		DownloadRate: s.currentDownloadRate,
		ActiveFlows:  len(s.activeFlows),
	}
	if len(s.timeSeries) >= s.opts.MaxTimeSeriesPts {
		s.timeSeries = append(s.timeSeries[1:], tsPoint)
	} else {
		s.timeSeries = append(s.timeSeries, tsPoint)
	}
}

func (s *ConnectionStore) recordRule(flow *domain.Flow, now time.Time) {
	if flow == nil || flow.Rule == "" {
		return
	}
	r, exists := s.rules[flow.Rule]
	if !exists {
		rType := "Match"
		if strings.HasPrefix(flow.Rule, "geosite") {
			rType = "Geosite"
		} else if strings.HasPrefix(flow.Rule, "geoip") {
			rType = "GeoIP"
		} else if strings.HasPrefix(flow.Rule, "protocol") {
			rType = "Protocol"
		} else if strings.HasPrefix(flow.Rule, "rule_set") {
			rType = "RuleSet"
		} else if strings.HasPrefix(flow.Rule, "domain") {
			rType = "Domain"
		} else if strings.HasPrefix(flow.Rule, "ip") {
			rType = "IP"
		} else if flow.Rule == "final" || flow.Rule == "default" {
			rType = "Default"
		}
		r = &domain.RuleInfo{
			Type:       rType,
			Payload:    flow.Rule,
			Proxy:      flow.Outbound,
			HitCount:   0,
			TotalBytes: 0,
			LastHitAt:  now.UnixMilli(),
			UUID:       fmt.Sprintf("rule-%d", len(s.rules)+1),
			Index:      len(s.rules) + 1,
		}
		s.rules[flow.Rule] = r
	}
	r.HitCount++
	r.TotalBytes += flow.UploadTotal + flow.DownloadTotal
	r.LastHitAt = now.UnixMilli()
	if flow.Outbound != "" {
		r.Proxy = flow.Outbound
	}
}

func (s *ConnectionStore) appendHistory(flow *domain.Flow) {
	if len(s.closedHistory) >= s.opts.MaxHistorySize {
		s.closedHistory = append(s.closedHistory[1:], flow)
	} else {
		s.closedHistory = append(s.closedHistory, flow)
	}
}

// MatchesInboundFilter tests if a flow satisfies the active inbound filter.
func MatchesInboundFilter(flow *domain.Flow, filter string) bool {
	if filter == "" || filter == "all" {
		return true
	}
	if filter == "tun" || filter == "tun:all" {
		return strings.EqualFold(flow.InboundType, "tun")
	}
	if strings.HasPrefix(filter, "tag:") {
		tag := strings.TrimPrefix(filter, "tag:")
		return flow.Inbound == tag
	}
	return flow.Inbound == filter || strings.EqualFold(flow.InboundType, filter)
}

// QueryOptions represents query parameters for retrieving flows.
type QueryOptions struct {
	Search     string `json:"search"`
	Process    string `json:"process"`
	Inbound    string `json:"inbound"`
	Outbound   string `json:"outbound"`
	Protocol   string `json:"protocol"`
	Network    string `json:"network"`
	ActiveOnly bool   `json:"activeOnly"`
	TUNOnly    bool   `json:"tunOnly"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
	SortBy     string `json:"sortBy"`
	SortDesc   bool   `json:"sortDesc"`
}

// FlowListResult contains paginated query results.
type FlowListResult struct {
	Flows      []*domain.Flow `json:"flows"`
	TotalCount int            `json:"totalCount"`
}

// GetFlows returns flows matching QueryOptions.
func (s *ConnectionStore) GetFlows(opts QueryOptions) FlowListResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	inboundFilter := s.selectedInboundFilter
	if opts.TUNOnly {
		inboundFilter = "tun:all"
	} else if opts.Inbound != "" {
		inboundFilter = opts.Inbound
	}

	searchLower := strings.ToLower(opts.Search)
	candidates := make([]*domain.Flow, 0, len(s.activeFlows)+len(s.closedHistory))

	for _, f := range s.activeFlows {
		if !MatchesInboundFilter(f, inboundFilter) {
			continue
		}
		if opts.Process != "" && (f.Process == nil || f.Process.ProcessName != opts.Process) {
			continue
		}
		if opts.Outbound != "" && f.Outbound != opts.Outbound {
			continue
		}
		if opts.Protocol != "" && !strings.EqualFold(f.Protocol, opts.Protocol) {
			continue
		}
		if opts.Network != "" && !strings.EqualFold(f.Network, opts.Network) {
			continue
		}
		if searchLower != "" && !matchesSearch(f, searchLower) {
			continue
		}
		candidates = append(candidates, f)
	}

	if !opts.ActiveOnly {
		for i := len(s.closedHistory) - 1; i >= 0; i-- {
			f := s.closedHistory[i]
			if !MatchesInboundFilter(f, inboundFilter) {
				continue
			}
			if opts.Process != "" && (f.Process == nil || f.Process.ProcessName != opts.Process) {
				continue
			}
			if opts.Outbound != "" && f.Outbound != opts.Outbound {
				continue
			}
			if opts.Protocol != "" && !strings.EqualFold(f.Protocol, opts.Protocol) {
				continue
			}
			if opts.Network != "" && !strings.EqualFold(f.Network, opts.Network) {
				continue
			}
			if searchLower != "" && !matchesSearch(f, searchLower) {
				continue
			}
			candidates = append(candidates, f)
		}
	}

	totalCount := len(candidates)
	sortFlows(candidates, opts.SortBy, opts.SortDesc)

	start := opts.Offset
	if start < 0 {
		start = 0
	}
	if start > totalCount {
		start = totalCount
	}
	end := totalCount
	if opts.Limit > 0 && start+opts.Limit < totalCount {
		end = start + opts.Limit
	}

	return FlowListResult{
		Flows:      candidates[start:end],
		TotalCount: totalCount,
	}
}

func matchesSearch(f *domain.Flow, q string) bool {
	if strings.Contains(strings.ToLower(f.Domain), q) {
		return true
	}
	if strings.Contains(strings.ToLower(f.Destination), q) {
		return true
	}
	if strings.Contains(strings.ToLower(f.Source), q) {
		return true
	}
	if strings.Contains(strings.ToLower(f.Rule), q) {
		return true
	}
	if strings.Contains(strings.ToLower(f.Outbound), q) {
		return true
	}
	if f.Process != nil {
		if strings.Contains(strings.ToLower(f.Process.ProcessName), q) {
			return true
		}
		if strings.Contains(strings.ToLower(f.Process.ProcessPath), q) {
			return true
		}
	}
	return false
}

func sortFlows(flows []*domain.Flow, sortBy string, sortDesc bool) {
	if len(flows) <= 1 || sortBy == "" {
		return
	}
	var less func(i, j int) bool
	switch sortBy {
	case "uploadRate":
		less = func(i, j int) bool { return flows[i].UploadRate < flows[j].UploadRate }
	case "downloadRate":
		less = func(i, j int) bool { return flows[i].DownloadRate < flows[j].DownloadRate }
	case "uploadTotal":
		less = func(i, j int) bool { return flows[i].UploadTotal < flows[j].UploadTotal }
	case "downloadTotal":
		less = func(i, j int) bool { return flows[i].DownloadTotal < flows[j].DownloadTotal }
	case "totalBytes":
		less = func(i, j int) bool {
			return (flows[i].UploadTotal + flows[i].DownloadTotal) < (flows[j].UploadTotal + flows[j].DownloadTotal)
		}
	case "createdAt":
		less = func(i, j int) bool { return flows[i].CreatedAt.Before(flows[j].CreatedAt) }
	case "domain":
		less = func(i, j int) bool { return flows[i].Domain < flows[j].Domain }
	case "process":
		less = func(i, j int) bool {
			p1, p2 := "", ""
			if flows[i].Process != nil {
				p1 = flows[i].Process.ProcessName
			}
			if flows[j].Process != nil {
				p2 = flows[j].Process.ProcessName
			}
			return p1 < p2
		}
	default:
		return
	}

	quickSort(flows, 0, len(flows)-1, less, sortDesc)
}

func quickSort(flows []*domain.Flow, low, high int, less func(i, j int) bool, desc bool) {
	if low < high {
		p := partition(flows, low, high, less, desc)
		quickSort(flows, low, p-1, less, desc)
		quickSort(flows, p+1, high, less, desc)
	}
}

func partition(flows []*domain.Flow, low, high int, less func(i, j int) bool, desc bool) int {
	pivotIdx := high
	i := low - 1
	for j := low; j < high; j++ {
		isLess := less(j, pivotIdx)
		condition := isLess
		if desc {
			condition = !isLess && (flows[j] != flows[pivotIdx])
		}
		if condition {
			i++
			flows[i], flows[j] = flows[j], flows[i]
		}
	}
	flows[i+1], flows[high] = flows[high], flows[i+1]
	return i + 1
}

// GetOverviewSummary returns real-time KPI metrics and Top-N aggregates.
func (s *ConnectionStore) GetOverviewSummary() domain.OverviewSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	inboundFilter := s.selectedInboundFilter
	var activeTUNCount, activeTotalCount, tcpCount, udpCount int

	processBytes := make(map[string]*domain.NamedAggregate)
	domainBytes := make(map[string]*domain.NamedAggregate)
	destBytes := make(map[string]*domain.NamedAggregate)
	outboundBytes := make(map[string]*domain.NamedAggregate)

	accumulate := func(f *domain.Flow) {
		isTUN := strings.EqualFold(f.InboundType, "tun")
		if isTUN {
			activeTUNCount++
		}
		activeTotalCount++

		if strings.EqualFold(f.Network, "tcp") {
			tcpCount++
		} else if strings.EqualFold(f.Network, "udp") {
			udpCount++
		}

		total := f.UploadTotal + f.DownloadTotal

		// Process
		pName := "Unknown"
		if f.Process != nil && f.Process.ProcessName != "" && f.Process.ProcessName != "Unknown" {
			pName = f.Process.ProcessName
		} else if f.Process != nil && f.Process.ProcessPath != "" {
			pName = f.Process.ProcessPath
		}
		if agg, ok := processBytes[pName]; ok {
			agg.TotalBytes += total
			agg.UploadTotal += f.UploadTotal
			agg.DownloadTotal += f.DownloadTotal
			agg.ActiveCount++
		} else {
			processBytes[pName] = &domain.NamedAggregate{
				Key:           pName,
				Name:          pName,
				TotalBytes:    total,
				UploadTotal:   f.UploadTotal,
				DownloadTotal: f.DownloadTotal,
				ActiveCount:   1,
			}
		}

		// Domain
		if f.Domain != "" {
			if agg, ok := domainBytes[f.Domain]; ok {
				agg.TotalBytes += total
				agg.ActiveCount++
			} else {
				domainBytes[f.Domain] = &domain.NamedAggregate{
					Key:         f.Domain,
					Name:        f.Domain,
					TotalBytes:  total,
					ActiveCount: 1,
				}
			}
		}

		// Destination
		if f.Destination != "" {
			if agg, ok := destBytes[f.Destination]; ok {
				agg.TotalBytes += total
				agg.ActiveCount++
			} else {
				destBytes[f.Destination] = &domain.NamedAggregate{
					Key:         f.Destination,
					Name:        f.Destination,
					TotalBytes:  total,
					ActiveCount: 1,
				}
			}
		}

		// Outbound
		if f.Outbound != "" {
			if agg, ok := outboundBytes[f.Outbound]; ok {
				agg.TotalBytes += total
				agg.ActiveCount++
			} else {
				outboundBytes[f.Outbound] = &domain.NamedAggregate{
					Key:         f.Outbound,
					Name:        f.Outbound,
					Category:    f.OutboundType,
					TotalBytes:  total,
					ActiveCount: 1,
				}
			}
		}
	}

	for _, f := range s.activeFlows {
		if MatchesInboundFilter(f, inboundFilter) {
			accumulate(f)
		}
	}

	topProcess := findMaxAggregate(processBytes)
	topDomain := findMaxAggregate(domainBytes)
	topDest := findMaxAggregate(destBytes)
	topOutbound := findMaxAggregate(outboundBytes)

	tsCopy := make([]domain.TimeSeriesPoint, len(s.timeSeries))
	copy(tsCopy, s.timeSeries)

	return domain.OverviewSummary{
		UploadRate:       s.currentUploadRate,
		DownloadRate:     s.currentDownloadRate,
		SessionUpload:    s.sessionUploadBytes,
		SessionDownload:  s.sessionDownloadBytes,
		ActiveTUNFlows:   activeTUNCount,
		ActiveTotalFlows: activeTotalCount,
		TCPCount:         tcpCount,
		UDPCount:         udpCount,
		TopProcess:       topProcess,
		TopDomain:        topDomain,
		TopDestination:   topDest,
		TopOutbound:      topOutbound,
		TimeSeries:       tsCopy,
	}
}

func findMaxAggregate(m map[string]*domain.NamedAggregate) *domain.NamedAggregate {
	var maxAgg *domain.NamedAggregate
	var maxVal int64 = -1
	for _, agg := range m {
		if agg.TotalBytes > maxVal {
			maxVal = agg.TotalBytes
			maxAgg = agg
		}
	}
	return maxAgg
}

// GetDiscoveredRules returns all rules discovered from live connections.
func (s *ConnectionStore) GetDiscoveredRules() []domain.RuleInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]domain.RuleInfo, 0, len(s.rules))
	for _, r := range s.rules {
		res = append(res, *r)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].HitCount > res[j].HitCount
	})

	return res
}

// GetMetadataCatalogs returns the list of known inbounds, outbounds, processes, and protocols.
func (s *ConnectionStore) GetMetadataCatalogs() (inbounds []string, outbounds []string, processes []string, protocols []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for k := range s.inbounds {
		inbounds = append(inbounds, k)
	}
	for k := range s.outbounds {
		outbounds = append(outbounds, k)
	}
	for k := range s.processes {
		processes = append(processes, k)
	}
	for k := range s.protocols {
		protocols = append(protocols, k)
	}
	return
}

// GetAllFlowsSnapshot returns a shallow clone of all active and history flows for batch Rust analytics.
func (s *ConnectionStore) GetAllFlowsSnapshot() []*domain.Flow {
	s.mu.RLock()
	defer s.mu.RUnlock()

	flows := make([]*domain.Flow, 0, len(s.activeFlows)+len(s.closedHistory))
	for _, f := range s.activeFlows {
		flows = append(flows, f)
	}
	flows = append(flows, s.closedHistory...)
	return flows
}
