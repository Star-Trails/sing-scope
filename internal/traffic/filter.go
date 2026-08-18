package traffic

import (
	"strings"

	"sing-scope/internal/domain"
)

// InboundFilterType defines the mode of inbound filtering.
type InboundFilterType string

const (
	FilterAllInbounds InboundFilterType = "all"
	FilterAllTUN      InboundFilterType = "tun"
	FilterSpecificTag InboundFilterType = "tag"
)

// InboundFilter specifies matching criteria for inbounds.
type InboundFilter struct {
	Type InboundFilterType `json:"type"`
	Tag  string            `json:"tag,omitempty"`
}

// Matches checks if a flow passes the inbound filter.
func (f InboundFilter) Matches(flow *domain.Flow) bool {
	if flow == nil {
		return false
	}
	switch f.Type {
	case FilterAllInbounds, "":
		return true
	case FilterAllTUN:
		return strings.EqualFold(flow.InboundType, "tun")
	case FilterSpecificTag:
		return flow.Inbound == f.Tag
	default:
		return true
	}
}

// FilterFlows applies the filter to a slice of flows.
func FilterFlows(flows []*domain.Flow, filter InboundFilter) []*domain.Flow {
	result := make([]*domain.Flow, 0, len(flows))
	for _, f := range flows {
		if filter.Matches(f) {
			result = append(result, f)
		}
	}
	return result
}
