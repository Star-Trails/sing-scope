package traffic

import (
	"testing"

	"sing-scope/internal/domain"
)

func TestInboundFilter_Matches(t *testing.T) {
	tunFlow := &domain.Flow{
		Inbound:     "tun-in",
		InboundType: "tun",
	}
	mixedFlow := &domain.Flow{
		Inbound:     "mixed-in",
		InboundType: "mixed",
	}

	filterAll := InboundFilter{Type: FilterAllInbounds}
	filterTUN := InboundFilter{Type: FilterAllTUN}
	filterTag := InboundFilter{Type: FilterSpecificTag, Tag: "mixed-in"}

	if !filterAll.Matches(tunFlow) || !filterAll.Matches(mixedFlow) {
		t.Error("FilterAllInbounds failed to match all flows")
	}

	if !filterTUN.Matches(tunFlow) || filterTUN.Matches(mixedFlow) {
		t.Error("FilterAllTUN matched non-TUN flow or missed TUN flow")
	}

	if filterTag.Matches(tunFlow) || !filterTag.Matches(mixedFlow) {
		t.Error("FilterSpecificTag failed tag check")
	}
}
