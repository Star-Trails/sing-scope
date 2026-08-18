package store

import (
	"fmt"
	"testing"
	"time"

	"sing-scope/internal/domain"
)

func TestConnectionStore_Lifecycle(t *testing.T) {
	s := NewConnectionStore(DefaultStoreOptions())

	// 1. Initial NEW event
	flow1 := &domain.Flow{
		ID:          "conn-1",
		Inbound:     "tun-in",
		InboundType: "tun",
		Network:     "tcp",
		Source:      "192.168.1.10:1001",
		Destination: "1.1.1.1:443",
		Domain:      "cloudflare.com",
		Outbound:    "proxy-hk",
		Process: &domain.ProcessInfo{
			ProcessName: "chrome.exe",
		},
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	s.ProcessBatch([]domain.FlowEvent{
		{
			Type:      domain.FlowEventNew,
			ID:        "conn-1",
			Flow:      flow1,
			Timestamp: time.Now(),
		},
	}, false)

	res := s.GetFlows(QueryOptions{})
	if res.TotalCount != 1 {
		t.Fatalf("expected 1 flow, got %d", res.TotalCount)
	}
	if res.Flows[0].ID != "conn-1" {
		t.Errorf("expected conn-1, got %s", res.Flows[0].ID)
	}

	// 2. UPDATE event with delta
	time.Sleep(50 * time.Millisecond)
	s.ProcessBatch([]domain.FlowEvent{
		{
			Type:          domain.FlowEventUpdate,
			ID:            "conn-1",
			UplinkDelta:   1000,
			DownlinkDelta: 5000,
			Timestamp:     time.Now(),
		},
	}, false)

	res = s.GetFlows(QueryOptions{})
	if res.TotalCount != 1 {
		t.Fatalf("expected 1 flow, got %d", res.TotalCount)
	}
	if res.Flows[0].UploadTotal != 1000 || res.Flows[0].DownloadTotal != 5000 {
		t.Errorf("unexpected totals: up=%d, down=%d", res.Flows[0].UploadTotal, res.Flows[0].DownloadTotal)
	}
	if res.Flows[0].DownloadRate <= 0 {
		t.Errorf("expected positive download rate, got %f", res.Flows[0].DownloadRate)
	}

	// 3. CLOSED event
	now := time.Now()
	s.ProcessBatch([]domain.FlowEvent{
		{
			Type:      domain.FlowEventClosed,
			ID:        "conn-1",
			ClosedAt:  &now,
			Timestamp: now,
		},
	}, false)

	// Active only query
	activeRes := s.GetFlows(QueryOptions{ActiveOnly: true})
	if activeRes.TotalCount != 0 {
		t.Errorf("expected 0 active flows, got %d", activeRes.TotalCount)
	}

	// Include closed history
	allRes := s.GetFlows(QueryOptions{ActiveOnly: false})
	if allRes.TotalCount != 1 {
		t.Errorf("expected 1 historical flow, got %d", allRes.TotalCount)
	}
	if allRes.Flows[0].IsActive {
		t.Error("expected historical flow to have IsActive == false")
	}

	// 4. RESET event
	flow2 := &domain.Flow{
		ID:          "conn-2",
		Inbound:     "mixed-in",
		InboundType: "mixed",
		Network:     "udp",
		Source:      "192.168.1.10:1002",
		Destination: "8.8.8.8:53",
		Domain:      "dns.google",
		Outbound:    "direct",
		CreatedAt:   time.Now(),
		IsActive:    true,
	}
	s.ProcessBatch([]domain.FlowEvent{
		{
			Type:      domain.FlowEventNew,
			ID:        "conn-2",
			Flow:      flow2,
			Timestamp: time.Now(),
		},
	}, true) // reset: true

	activeRes = s.GetFlows(QueryOptions{ActiveOnly: true})
	if activeRes.TotalCount != 1 || activeRes.Flows[0].ID != "conn-2" {
		t.Errorf("expected reset to leave only conn-2 active, got %d flows", activeRes.TotalCount)
	}
}

func TestConnectionStore_TUNFiltering(t *testing.T) {
	s := NewConnectionStore(DefaultStoreOptions())

	fTun := &domain.Flow{
		ID:          "tun-flow",
		Inbound:     "tun-in",
		InboundType: "tun",
		Network:     "tcp",
		CreatedAt:   time.Now(),
		IsActive:    true,
	}
	fMixed := &domain.Flow{
		ID:          "mixed-flow",
		Inbound:     "mixed-in",
		InboundType: "mixed",
		Network:     "tcp",
		CreatedAt:   time.Now(),
		IsActive:    true,
	}

	s.ProcessBatch([]domain.FlowEvent{
		{Type: domain.FlowEventNew, ID: fTun.ID, Flow: fTun},
		{Type: domain.FlowEventNew, ID: fMixed.ID, Flow: fMixed},
	}, false)

	// All inbounds
	s.SetInboundFilter("all")
	res := s.GetFlows(QueryOptions{})
	if res.TotalCount != 2 {
		t.Errorf("expected 2 flows with filter 'all', got %d", res.TotalCount)
	}

	// TUN only
	s.SetInboundFilter("tun:all")
	res = s.GetFlows(QueryOptions{})
	if res.TotalCount != 1 || res.Flows[0].ID != "tun-flow" {
		t.Errorf("expected 1 flow with filter 'tun:all', got %d", res.TotalCount)
	}

	// Tag specific
	s.SetInboundFilter("tag:mixed-in")
	res = s.GetFlows(QueryOptions{})
	if res.TotalCount != 1 || res.Flows[0].ID != "mixed-flow" {
		t.Errorf("expected 1 flow with tag:mixed-in, got %d", res.TotalCount)
	}
}

func TestConnectionStore_OverviewSummary(t *testing.T) {
	s := NewConnectionStore(DefaultStoreOptions())

	for i := range 10 {
		flow := &domain.Flow{
			ID:            fmt.Sprintf("conn-%d", i),
			Inbound:       "tun-in",
			InboundType:   "tun",
			Network:       "tcp",
			Domain:        fmt.Sprintf("domain-%d.com", i%3),
			Destination:   fmt.Sprintf("10.0.0.%d:443", i),
			Outbound:      fmt.Sprintf("node-%d", i%2),
			OutboundType:  "shadowsocks",
			UploadTotal:   int64(i * 100),
			DownloadTotal: int64(i * 500),
			Process: &domain.ProcessInfo{
				ProcessName: fmt.Sprintf("proc-%d.exe", i%2),
			},
			CreatedAt: time.Now(),
			IsActive:  true,
		}
		s.ProcessBatch([]domain.FlowEvent{
			{Type: domain.FlowEventNew, ID: flow.ID, Flow: flow},
		}, false)
	}

	summary := s.GetOverviewSummary()
	if summary.ActiveTUNFlows != 10 {
		t.Errorf("expected 10 active TUN flows, got %d", summary.ActiveTUNFlows)
	}
	if summary.TopProcess == nil || summary.TopProcess.Name != "proc-1.exe" {
		t.Errorf("expected TopProcess proc-1.exe, got %+v", summary.TopProcess)
	}
}
