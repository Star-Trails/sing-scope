package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sing-scope/internal/domain"
	"sing-scope/internal/singboxapi"
	"sing-scope/internal/store"
)

func main() {
	serverURL := flag.String("url", "http://127.0.0.1:9090", "sing-box API server URL (e.g. http://127.0.0.1:9090)")
	secret := flag.String("secret", "", "sing-box API secret / bearer token")
	tunOnly := flag.Bool("tun-only", false, "Filter to TUN traffic only")
	flag.Parse()

	fmt.Println("==================================================")
	fmt.Println("  sing-scope CLI Traffic Collector & Debugger")
	fmt.Println("==================================================")
	fmt.Printf("Connecting to sing-box API at: %s\n", *serverURL)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	connStore := store.NewConnectionStore(store.DefaultStoreOptions())
	if *tunOnly {
		connStore.SetInboundFilter("tun:all")
		fmt.Println("Filtering: TUN only")
	}

	mgr := singboxapi.NewManager(
		singboxapi.ClientOptions{
			ServerURL: *serverURL,
			Secret:    *secret,
			Timeout:   5 * time.Second,
		},
		singboxapi.DefaultReconnectOptions(),
		func(info *domain.ServerConnectionInfo) {
			fmt.Printf("[STATUS] State: %s | Sing-box: %s (API v%d) | Error: %s\n",
				info.State, info.SingBoxVersion, info.APIVersion, info.ErrorMessage)
		},
		func(events []domain.FlowEvent, isReset bool) {
			if isReset {
				fmt.Println("[STREAM] Received snapshot RESET (clearing stale state)")
			}
			connStore.ProcessBatch(events, isReset)

			for _, e := range events {
				switch e.Type {
				case domain.FlowEventNew:
					if e.Flow != nil {
						fmt.Printf("[NEW] Flow %s: %s://%s -> %s (inbound: %s/%s, outbound: %s/%s, domain: %s)\n",
							e.Flow.ID[:8], e.Flow.Network, e.Flow.Source, e.Flow.Destination,
							e.Flow.Inbound, e.Flow.InboundType, e.Flow.Outbound, e.Flow.OutboundType,
							e.Flow.Domain)
					}
				case domain.FlowEventUpdate:
					fmt.Printf("[UPDATE] Flow %s: ▲ +%s (up) | ▼ +%s (down)\n",
						e.ID[:8], formatBytes(e.UplinkDelta), formatBytes(e.DownlinkDelta))
				case domain.FlowEventClosed:
					fmt.Printf("[CLOSED] Flow %s closed at %v\n", e.ID[:8], e.ClosedAt)
				}
			}
		},
		func(status *domain.SystemStatus) {
			fmt.Printf("[METRICS] Memory: %s | Goroutines: %d | Conns In/Out: %d/%d | Up: %s/s | Down: %s/s\n",
				formatBytes(int64(status.Memory)), status.Goroutines, status.ConnectionsIn, status.ConnectionsOut,
				formatBytes(status.Uplink), formatBytes(status.Downlink))
		},
		func(logs []domain.LogMessage) {
			for _, l := range logs {
				fmt.Printf("[LOG:%s] %s\n", l.Level, l.Message)
			}
		},
		func(groups []domain.OutboundGroup) {
			fmt.Printf("[GROUPS] Received %d outbound groups\n", len(groups))
		},
	)

	mgr.Start(ctx)
	defer mgr.Stop()

	// Periodic console summary reporter
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nCollector stopped by user.")
			return
		case <-ticker.C:
			summary := connStore.GetOverviewSummary()
			fmt.Printf("\n--- [SUMMARY] Active Flows: %d (TUN: %d) | Rate: ▲ %s/s, ▼ %s/s | Session: ▲ %s, ▼ %s ---\n",
				summary.ActiveTotalFlows, summary.ActiveTUNFlows,
				formatBytes(int64(summary.UploadRate)), formatBytes(int64(summary.DownloadRate)),
				formatBytes(summary.SessionUpload), formatBytes(summary.SessionDownload))
		}
	}
}


func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
