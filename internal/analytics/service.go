package analytics

import (
	"encoding/json"
	"sort"
	"time"

	"sing-scope/internal/domain"
	"sing-scope/internal/ffi"
	"sing-scope/internal/store"
)

// BatchAnalysisResult represents the complete analytics breakdown.
type BatchAnalysisResult struct {
	TotalFlows         int                       `json:"totalFlows"`
	ActiveFlows        int                       `json:"activeFlows"`
	TotalUploadBytes   int64                     `json:"totalUploadBytes"`
	TotalDownloadBytes int64                     `json:"totalDownloadBytes"`
	TotalUploadRate    float64                   `json:"totalUploadRate"`
	TotalDownloadRate  float64                   `json:"totalDownloadRate"`
	ByProcess          []domain.ProcessAggregate `json:"byProcess"`
	ByDomain           []domain.NamedAggregate   `json:"byDomain"`
	ByDestination      []domain.NamedAggregate   `json:"byDestination"`
	ByOutbound         []domain.NamedAggregate   `json:"byOutbound"`
	ByRule             []domain.NamedAggregate   `json:"byRule"`
	ByProtocol         []domain.NamedAggregate   `json:"byProtocol"`
	ComputeTimeUs      uint64                    `json:"computeTimeUs"`
	Engine             string                    `json:"engine"` // "rust-native" or "pure-go"
}

// Service provides high-level analytics operations over flows.
type Service struct {
	store *store.ConnectionStore
}

// NewService creates a new Analytics Service.
func NewService(s *store.ConnectionStore) *Service {
	return &Service{store: s}
}

// AnalyzeBatch computes full multi-dimensional analytics for current flows.
func (s *Service) AnalyzeBatch(inboundFilter string, topN int) (BatchAnalysisResult, error) {
	if topN <= 0 {
		topN = 15
	}

	flows := s.store.GetAllFlowsSnapshot()

	// Try Rust native FFI first if available
	if ffi.IsNativeCoreAvailable() {
		req := map[string]any{
			"flows":         flows,
			"topN":          topN,
			"inboundFilter": inboundFilter,
		}

		outBytes, err := ffi.NativeAnalyzeBatch(req)
		if err == nil && len(outBytes) > 0 {
			var res BatchAnalysisResult
			if err := json.Unmarshal(outBytes, &res); err == nil && res.TotalFlows >= 0 {
				res.Engine = "rust-native (" + ffi.GetCoreVersion() + ")"
				return res, nil
			}
		}
	}

	// Fallback to Pure Go Analytics Engine
	return s.analyzeBatchPureGo(flows, inboundFilter, topN), nil
}

func (s *Service) analyzeBatchPureGo(flows []*domain.Flow, filter string, topN int) BatchAnalysisResult {
	start := time.Now()

	var totalFlows, activeFlows int
	var totalUpload, totalDownload int64
	var totalUpRate, totalDownRate float64

	domainMap := make(map[string]*domain.NamedAggregate)
	destMap := make(map[string]*domain.NamedAggregate)
	outboundMap := make(map[string]*domain.NamedAggregate)
	ruleMap := make(map[string]*domain.NamedAggregate)
	protocolMap := make(map[string]*domain.NamedAggregate)

	type procAcc struct {
		name          string
		path          string
		pid           uint32
		connCount     int
		activeCount   int
		uploadTotal   int64
		downloadTotal int64
		uploadRate    float64
		downloadRate  float64
		domains       map[string]int64
		destinations  map[string]int64
	}
	procMap := make(map[string]*procAcc)

	for _, f := range flows {
		if !store.MatchesInboundFilter(f, filter) {
			continue
		}

		totalFlows++
		if f.IsActive {
			activeFlows++
			totalUpRate += f.UploadRate
			totalDownRate += f.DownloadRate
		}
		totalUpload += f.UploadTotal
		totalDownload += f.DownloadTotal

		totalBytes := f.UploadTotal + f.DownloadTotal
		totalRate := f.UploadRate + f.DownloadRate

		// Domain
		if f.Domain != "" {
			entry, ok := domainMap[f.Domain]
			if !ok {
				entry = &domain.NamedAggregate{Key: f.Domain, Name: f.Domain}
				domainMap[f.Domain] = entry
			}
			entry.ConnectionCount++
			if f.IsActive {
				entry.ActiveCount++
				entry.UploadRate += f.UploadRate
				entry.DownloadRate += f.DownloadRate
				entry.TotalRate += totalRate
			}
			entry.UploadTotal += f.UploadTotal
			entry.DownloadTotal += f.DownloadTotal
			entry.TotalBytes += totalBytes
		}

		// Destination
		if f.Destination != "" {
			entry, ok := destMap[f.Destination]
			if !ok {
				entry = &domain.NamedAggregate{Key: f.Destination, Name: f.Destination}
				destMap[f.Destination] = entry
			}
			entry.ConnectionCount++
			if f.IsActive {
				entry.ActiveCount++
				entry.UploadRate += f.UploadRate
				entry.DownloadRate += f.DownloadRate
				entry.TotalRate += totalRate
			}
			entry.UploadTotal += f.UploadTotal
			entry.DownloadTotal += f.DownloadTotal
			entry.TotalBytes += totalBytes
		}

		// Outbound
		if f.Outbound != "" {
			entry, ok := outboundMap[f.Outbound]
			if !ok {
				entry = &domain.NamedAggregate{Key: f.Outbound, Name: f.Outbound, Category: f.OutboundType}
				outboundMap[f.Outbound] = entry
			}
			entry.ConnectionCount++
			if f.IsActive {
				entry.ActiveCount++
				entry.UploadRate += f.UploadRate
				entry.DownloadRate += f.DownloadRate
				entry.TotalRate += totalRate
			}
			entry.UploadTotal += f.UploadTotal
			entry.DownloadTotal += f.DownloadTotal
			entry.TotalBytes += totalBytes
		}

		// Rule
		if f.Rule != "" {
			entry, ok := ruleMap[f.Rule]
			if !ok {
				entry = &domain.NamedAggregate{Key: f.Rule, Name: f.Rule}
				ruleMap[f.Rule] = entry
			}
			entry.ConnectionCount++
			if f.IsActive {
				entry.ActiveCount++
				entry.UploadRate += f.UploadRate
				entry.DownloadRate += f.DownloadRate
				entry.TotalRate += totalRate
			}
			entry.UploadTotal += f.UploadTotal
			entry.DownloadTotal += f.DownloadTotal
			entry.TotalBytes += totalBytes
		}

		// Protocol
		if f.Protocol != "" {
			entry, ok := protocolMap[f.Protocol]
			if !ok {
				entry = &domain.NamedAggregate{Key: f.Protocol, Name: f.Protocol}
				protocolMap[f.Protocol] = entry
			}
			entry.ConnectionCount++
			if f.IsActive {
				entry.ActiveCount++
				entry.UploadRate += f.UploadRate
				entry.DownloadRate += f.DownloadRate
				entry.TotalRate += totalRate
			}
			entry.UploadTotal += f.UploadTotal
			entry.DownloadTotal += f.DownloadTotal
			entry.TotalBytes += totalBytes
		}

		// Process
		pName := "Unknown"
		pPath := ""
		var pID uint32
		if f.Process != nil {
			if f.Process.ProcessName != "" {
				pName = f.Process.ProcessName
			}
			pPath = f.Process.ProcessPath
			pID = f.Process.ProcessID
		}

		pEntry, ok := procMap[pName]
		if !ok {
			pEntry = &procAcc{
				name:         pName,
				path:         pPath,
				pid:          pID,
				domains:      make(map[string]int64),
				destinations: make(map[string]int64),
			}
			procMap[pName] = pEntry
		}
		pEntry.connCount++
		if f.IsActive {
			pEntry.activeCount++
			pEntry.uploadRate += f.UploadRate
			pEntry.downloadRate += f.DownloadRate
		}
		pEntry.uploadTotal += f.UploadTotal
		pEntry.downloadTotal += f.DownloadTotal
		if f.Domain != "" {
			pEntry.domains[f.Domain] += totalBytes
		}
		if f.Destination != "" {
			pEntry.destinations[f.Destination] += totalBytes
		}
	}

	// Sort slices
	byDomain := sortAndTruncateMap(domainMap, topN)
	byDest := sortAndTruncateMap(destMap, topN)
	byOutbound := sortAndTruncateMap(outboundMap, topN)
	byRule := sortAndTruncateMap(ruleMap, topN)
	byProtocol := sortAndTruncateMap(protocolMap, topN)

	byProcess := make([]domain.ProcessAggregate, 0, len(procMap))
	for _, p := range procMap {
		totalB := p.uploadTotal + p.downloadTotal

		topDoms := sortMapToAggregates(p.domains, 5)
		topDests := sortMapToAggregates(p.destinations, 5)

		byProcess = append(byProcess, domain.ProcessAggregate{
			ProcessName:     p.name,
			ProcessPath:     p.path,
			ProcessID:       p.pid,
			ConnectionCount: p.connCount,
			ActiveCount:     p.activeCount,
			UploadTotal:     p.uploadTotal,
			DownloadTotal:   p.downloadTotal,
			TotalBytes:      totalB,
			UploadRate:      p.uploadRate,
			DownloadRate:    p.downloadRate,
			TopDomains:      topDoms,
			TopDestinations: topDests,
		})
	}
	sort.Slice(byProcess, func(i, j int) bool {
		return byProcess[i].TotalBytes > byProcess[j].TotalBytes
	})
	if len(byProcess) > topN {
		byProcess = byProcess[:topN]
	}

	elapsedUs := uint64(time.Since(start).Microseconds())

	return BatchAnalysisResult{
		TotalFlows:         totalFlows,
		ActiveFlows:        activeFlows,
		TotalUploadBytes:   totalUpload,
		TotalDownloadBytes: totalDownload,
		TotalUploadRate:    totalUpRate,
		TotalDownloadRate:  totalDownRate,
		ByProcess:          byProcess,
		ByDomain:           byDomain,
		ByDestination:      byDest,
		ByOutbound:         byOutbound,
		ByRule:             byRule,
		ByProtocol:         byProtocol,
		ComputeTimeUs:      elapsedUs,
		Engine:             "pure-go",
	}
}

func sortAndTruncateMap(m map[string]*domain.NamedAggregate, topN int) []domain.NamedAggregate {
	list := make([]domain.NamedAggregate, 0, len(m))
	for _, v := range m {
		list = append(list, *v)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].TotalBytes > list[j].TotalBytes
	})
	if len(list) > topN {
		list = list[:topN]
	}
	return list
}

func sortMapToAggregates(m map[string]int64, topN int) []domain.NamedAggregate {
	list := make([]domain.NamedAggregate, 0, len(m))
	for k, v := range m {
		list = append(list, domain.NamedAggregate{
			Key:        k,
			Name:       k,
			TotalBytes: v,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].TotalBytes > list[j].TotalBytes
	})
	if len(list) > topN {
		list = list[:topN]
	}
	return list
}
