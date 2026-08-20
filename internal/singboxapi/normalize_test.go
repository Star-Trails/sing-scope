package singboxapi

import (
	"testing"
	"time"

	"sing-scope/internal/domain"
	pb "sing-scope/internal/singboxapi/gen"
)


func TestNormalizeConnection(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	pbConn := &pb.Connection{
		Id:           "11111111-2222-3333-4444-555555555555",
		Inbound:      "tun-in",
		InboundType:  "tun",
		IpVersion:    4,
		Network:      "tcp",
		Source:       "192.168.1.100:54321",
		Destination:  "1.1.1.1:443",
		Domain:       "one.one.one.one",
		Protocol:     "tls",
		User:         "default",
		CreatedAt:    nowMs,
		Uplink:       1024,
		Downlink:     4096,
		UplinkTotal:  10240,
		DownlinkTotal: 40960,
		Rule:         "match-all",
		Outbound:     "direct",
		OutboundType: "direct",
	}

	flow := NormalizeConnection(pbConn)
	if flow == nil {
		t.Fatal("expected flow, got nil")
	}

	if flow.ID != pbConn.Id {
		t.Errorf("expected ID %s, got %s", pbConn.Id, flow.ID)
	}
	if flow.InboundType != "tun" {
		t.Errorf("expected InboundType 'tun', got %s", flow.InboundType)
	}
	if flow.UploadTotal != 10240 || flow.DownloadTotal != 40960 {
		t.Errorf("unexpected totals: up=%d, down=%d", flow.UploadTotal, flow.DownloadTotal)
	}
	if !flow.IsActive {
		t.Error("expected flow to be active")
	}
}

func TestNormalizeConnectionEvent(t *testing.T) {
	now := time.Now()
	pbEvt := &pb.ConnectionEvent{
		Type:          pb.ConnectionEventType_CONNECTION_EVENT_UPDATE,
		Id:            "test-id",
		UplinkDelta:   500,
		DownlinkDelta: 1500,
	}

	domEvt := NormalizeConnectionEvent(pbEvt, now)
	if domEvt == nil {
		t.Fatal("expected non-nil FlowEvent")
	}
	if domEvt.Type != domain.FlowEventUpdate {
		t.Errorf("expected UPDATE event, got %v", domEvt.Type)
	}
	if domEvt.UplinkDelta != 500 || domEvt.DownlinkDelta != 1500 {
		t.Errorf("unexpected deltas: up=%d, down=%d", domEvt.UplinkDelta, domEvt.DownlinkDelta)
	}
}
