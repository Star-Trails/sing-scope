package store

import (
	"fmt"
	"testing"
	"time"

	"sing-scope/internal/domain"
)

func BenchmarkConnectionStore_ProcessBatch_1000Flows(b *testing.B) {
	s := NewConnectionStore(DefaultStoreOptions())
	events := generateSyntheticEvents(1000)

	b.ResetTimer()
	for range b.N {
		s.ProcessBatch(events, false)
	}
}

func BenchmarkConnectionStore_ProcessBatch_10000Flows(b *testing.B) {
	s := NewConnectionStore(DefaultStoreOptions())
	events := generateSyntheticEvents(10000)

	b.ResetTimer()
	for range b.N {
		s.ProcessBatch(events, false)
	}
}

func BenchmarkConnectionStore_Query_10000Active(b *testing.B) {
	s := NewConnectionStore(DefaultStoreOptions())
	events := generateSyntheticEvents(10000)
	s.ProcessBatch(events, true)

	opts := QueryOptions{
		Search:   "domain-5",
		Limit:    50,
		SortBy:   "downloadRate",
		SortDesc: true,
	}

	b.ResetTimer()
	for range b.N {
		_ = s.GetFlows(opts)
	}
}

func BenchmarkConnectionStore_Overview_10000Active(b *testing.B) {
	s := NewConnectionStore(DefaultStoreOptions())
	events := generateSyntheticEvents(10000)
	s.ProcessBatch(events, true)

	b.ResetTimer()
	for range b.N {
		_ = s.GetOverviewSummary()
	}
}

func TestConnectionStore_Synthetic_100kHistoricalFlows(t *testing.T) {
	opts := StoreOptions{
		MaxHistorySize:   100000,
		MaxTimeSeriesPts: 900,
	}
	s := NewConnectionStore(opts)

	start := time.Now()
	// Insert 100,000 flows in batches of 5,000
	const totalFlows = 100000
	const batchSize = 5000

	for batchStart := 0; batchStart < totalFlows; batchStart += batchSize {
		batch := make([]domain.FlowEvent, batchSize)
		for j := range batchSize {
			idx := batchStart + j
			closedAt := time.Now()
			flow := &domain.Flow{
				ID:            fmt.Sprintf("flow-%d", idx),
				Inbound:       "tun-in",
				InboundType:   "tun",
				Network:       "tcp",
				Source:        "192.168.1.10:12345",
				Destination:   fmt.Sprintf("10.%d.%d.%d:443", (idx>>16)&255, (idx>>8)&255, idx&255),
				Domain:        fmt.Sprintf("sub-%d.example.com", idx%1000),
				Outbound:      fmt.Sprintf("node-%d", idx%10),
				OutboundType:  "vless",
				UploadTotal:   int64(idx * 10),
				DownloadTotal: int64(idx * 50),
				CreatedAt:     time.Now().Add(-10 * time.Minute),
				ClosedAt:      &closedAt,
				IsActive:      false,
			}
			batch[j] = domain.FlowEvent{
				Type:      domain.FlowEventNew,
				ID:        flow.ID,
				Flow:      flow,
				Timestamp: closedAt,
			}
		}
		s.ProcessBatch(batch, false)
	}

	elapsed := time.Since(start)
	t.Logf("Inserted 100,000 historical flows in %v", elapsed)

	// Query over 100,000 historical flows
	queryStart := time.Now()
	res := s.GetFlows(QueryOptions{
		Search:   "example.com",
		Limit:    100,
		SortBy:   "downloadTotal",
		SortDesc: true,
	})
	queryElapsed := time.Since(queryStart)

	t.Logf("Queried 100,000 historical flows in %v: found %d matches (returned top %d)",
		queryElapsed, res.TotalCount, len(res.Flows))

	if res.TotalCount != 100000 {
		t.Errorf("expected 100,000 matched historical flows, got %d", res.TotalCount)
	}
	if len(res.Flows) != 100 {
		t.Errorf("expected 100 paginated results, got %d", len(res.Flows))
	}
}

func generateSyntheticEvents(count int) []domain.FlowEvent {
	events := make([]domain.FlowEvent, count)
	now := time.Now()
	for i := range count {
		flow := &domain.Flow{
			ID:            fmt.Sprintf("flow-%d", i),
			Inbound:       "tun-in",
			InboundType:   "tun",
			Network:       "tcp",
			Source:        "192.168.1.5:54321",
			Destination:   fmt.Sprintf("1.1.1.%d:443", i%250),
			Domain:        fmt.Sprintf("domain-%d.org", i%100),
			Protocol:      "tls",
			Outbound:      fmt.Sprintf("proxy-%d", i%5),
			OutboundType:  "vless",
			UploadTotal:   int64(i * 100),
			DownloadTotal: int64(i * 500),
			CreatedAt: now,
			IsActive:  true,
		}
		events[i] = domain.FlowEvent{
			Type:      domain.FlowEventNew,
			ID:        flow.ID,
			Flow:      flow,
			Timestamp: now,
		}
	}
	return events
}
