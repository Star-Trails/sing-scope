package singboxapi

import (
	"time"

	"sing-scope/internal/domain"
	pb "sing-scope/internal/singboxapi/gen"
)

// NormalizeConnection converts protobuf Connection to domain Flow.
func NormalizeConnection(proto *pb.Connection) *domain.Flow {
	if proto == nil {
		return nil
	}

	var createdAt time.Time
	if proto.GetCreatedAt() > 0 {
		createdAt = time.UnixMilli(proto.GetCreatedAt())
	} else {
		createdAt = time.Now()
	}

	var closedAt *time.Time
	if proto.GetClosedAt() > 0 {
		t := time.UnixMilli(proto.GetClosedAt())
		closedAt = &t
	}

	return &domain.Flow{
		ID:            proto.GetId(),
		Inbound:       proto.GetInbound(),
		InboundType:   proto.GetInboundType(),
		IPVersion:     int(proto.GetIpVersion()),
		Network:       proto.GetNetwork(),
		Source:        proto.GetSource(),
		Destination:   proto.GetDestination(),
		Domain:        proto.GetDomain(),
		Protocol:      proto.GetProtocol(),
		User:          proto.GetUser(),
		FromOutbound:  proto.GetFromOutbound(),
		Rule:          proto.GetRule(),
		Outbound:      proto.GetOutbound(),
		OutboundType:  proto.GetOutboundType(),
		ChainList:     proto.GetChainList(),
		CreatedAt:     createdAt,
		ClosedAt:      closedAt,
		UploadTotal:   proto.GetUplinkTotal(),
		DownloadTotal: proto.GetDownlinkTotal(),
		UploadRate:    float64(proto.GetUplink()),
		DownloadRate:  float64(proto.GetDownlink()),
		LastActiveAt:  time.Now(),
		IsActive:      closedAt == nil,
	}
}

// NormalizeConnectionEvent converts protobuf ConnectionEvent to domain FlowEvent.
func NormalizeConnectionEvent(proto *pb.ConnectionEvent, now time.Time) *domain.FlowEvent {
	if proto == nil {
		return nil
	}

	var eventType domain.FlowEventType
	switch proto.GetType() {
	case pb.ConnectionEventType_CONNECTION_EVENT_NEW:
		eventType = domain.FlowEventNew
	case pb.ConnectionEventType_CONNECTION_EVENT_UPDATE:
		eventType = domain.FlowEventUpdate
	case pb.ConnectionEventType_CONNECTION_EVENT_CLOSED:
		eventType = domain.FlowEventClosed
	default:
		eventType = domain.FlowEventUpdate
	}

	var closedAt *time.Time
	if proto.GetClosedAt() > 0 {
		t := time.UnixMilli(proto.GetClosedAt())
		closedAt = &t
	}

	return &domain.FlowEvent{
		Type:          eventType,
		ID:            proto.GetId(),
		Flow:          NormalizeConnection(proto.GetConnection()),
		UplinkDelta:   proto.GetUplinkDelta(),
		DownlinkDelta: proto.GetDownlinkDelta(),
		ClosedAt:      closedAt,
		Timestamp:     now,
	}
}

// NormalizeStatus converts protobuf Status to domain SystemStatus.
func NormalizeStatus(proto *pb.Status, now time.Time) *domain.SystemStatus {
	if proto == nil {
		return nil
	}
	return &domain.SystemStatus{
		Memory:           proto.GetMemory(),
		Goroutines:       proto.GetGoroutines(),
		ConnectionsIn:    proto.GetConnectionsIn(),
		ConnectionsOut:   proto.GetConnectionsOut(),
		TrafficAvailable: proto.GetTrafficAvailable(),
		Uplink:           proto.GetUplink(),
		Downlink:         proto.GetDownlink(),
		UplinkTotal:      proto.GetUplinkTotal(),
		DownlinkTotal:    proto.GetDownlinkTotal(),
		Timestamp:        now,
	}
}

// NormalizeLogs converts protobuf Log batch to domain LogMessage slice.
func NormalizeLogs(proto *pb.Log, now time.Time) []domain.LogMessage {
	if proto == nil || len(proto.GetMessages()) == 0 {
		return nil
	}
	res := make([]domain.LogMessage, 0, len(proto.GetMessages()))
	for _, m := range proto.GetMessages() {
		var lvl domain.LogLevel
		switch m.GetLevel() {
		case pb.LogLevel_PANIC:
			lvl = domain.LogLevelPanic
		case pb.LogLevel_FATAL:
			lvl = domain.LogLevelFatal
		case pb.LogLevel_ERROR:
			lvl = domain.LogLevelError
		case pb.LogLevel_WARN:
			lvl = domain.LogLevelWarn
		case pb.LogLevel_INFO:
			lvl = domain.LogLevelInfo
		case pb.LogLevel_DEBUG:
			lvl = domain.LogLevelDebug
		case pb.LogLevel_TRACE:
			lvl = domain.LogLevelTrace
		default:
			lvl = domain.LogLevelInfo
		}
		res = append(res, domain.LogMessage{
			Level:     lvl,
			Message:   m.GetMessage(),
			Timestamp: now,
		})
	}
	return res
}

// NormalizeGroups converts protobuf Groups to domain OutboundGroup slice.
func NormalizeGroups(proto *pb.Groups) []domain.OutboundGroup {
	if proto == nil || len(proto.GetGroup()) == 0 {
		return nil
	}
	res := make([]domain.OutboundGroup, 0, len(proto.GetGroup()))
	for _, g := range proto.GetGroup() {
		items := make([]domain.GroupItem, 0, len(g.GetItems()))
		for _, it := range g.GetItems() {
			items = append(items, domain.GroupItem{
				Tag:          it.GetTag(),
				Type:         it.GetType(),
				URLTestTime:  it.GetUrlTestTime(),
				URLTestDelay: it.GetUrlTestDelay(),
			})
		}
		res = append(res, domain.OutboundGroup{
			Tag:        g.GetTag(),
			Type:       g.GetType(),
			Selectable: g.GetSelectable(),
			Selected:   g.GetSelected(),
			IsExpand:   g.GetIsExpand(),
			Items:      items,
		})
	}
	return res
}

// NormalizeOutboundList converts protobuf OutboundList to domain OutboundInfo slice.
func NormalizeOutboundList(proto *pb.OutboundList) []domain.OutboundInfo {
	if proto == nil || len(proto.GetOutbounds()) == 0 {
		return nil
	}
	res := make([]domain.OutboundInfo, 0, len(proto.GetOutbounds()))
	for _, it := range proto.GetOutbounds() {
		res = append(res, domain.OutboundInfo{
			Tag:          it.GetTag(),
			Type:         it.GetType(),
			URLTestTime:  it.GetUrlTestTime(),
			URLTestDelay: it.GetUrlTestDelay(),
		})
	}
	return res
}
