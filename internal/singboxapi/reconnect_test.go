package singboxapi

import (
	"testing"
	"time"
)

func TestReconnectOptions_Backoff(t *testing.T) {
	opts := DefaultReconnectOptions()

	var interval time.Duration
	for range 10 {
		interval = opts.NextInterval(interval)
		if interval < 100*time.Millisecond {
			t.Errorf("interval %v below minimum floor", interval)
		}
		if interval > 12*time.Second {
			t.Errorf("interval %v exceeded max upper bound with jitter", interval)
		}
	}
}

func TestCheckCompatibility(t *testing.T) {
	res := CheckCompatibility("1.14.0-beta.17", 4)
	if !res.Compatible || res.Degraded {
		t.Errorf("expected fully compatible, got %+v", res)
	}

	resIncompat := CheckCompatibility("1.10.0", 3)
	if resIncompat.Compatible {
		t.Errorf("expected incompatible for API v3, got %+v", resIncompat)
	}
}
