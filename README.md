# sing-scope: Cross-Platform sing-box Traffic Analyzer

A polished, high-performance cross-platform desktop application built for analyzing network connections, traffic flows, and bandwidth metrics exposed by the **new sing-box 1.14+ gRPC API (`StartedService`)**.

- **Author**: Star-Trails <startrails01@outlook.com>
- **License**: [MIT License](LICENSE)

---

## 1. Overview & Key Capabilities

- **Native Wails v3 Desktop Architecture**: Uses `github.com/wailsapp/wails/v3` to deliver a pure native WebView desktop application (`PE32+ GUI`), with no command line popups or external browser dependencies.
- **Strict New gRPC API Boundary**: Directly communicates with the sing-box 1.14+ API service (`services: type: api`) over native gRPC HTTP/2 / h2c / TLS with Bearer token authentication.
- **Authoritative Connection State Machine**: Tracks real-time flow lifecycle (`NEW`, `UPDATE` delta calculation, `CLOSED`, `reset: true` state re-anchoring).
- **TUN Traffic Attribution**: Identifies and displays the true client process origin (`chrome.exe`, `curl`, `Discord.exe`) rather than redundant virtual gateway IPs.
- **Rust Analytics Core**: Features coarse-grained C-ABI batch analytics (`rust/traffic-core`) for multi-dimensional flow aggregation, rule matching volume, and time-series downsampling.
- **Zashboard Visual Language**: Fully adopts the refined UI design, layout, and component architecture of [Zephyruso/zashboard](https://github.com/Zephyruso/zashboard).

---

## 2. Technology Stack

- **Desktop Framework**: Go + Wails v3 (`github.com/wailsapp/wails/v3@v3.0.0-beta.9`)
- **Backend & Networking**: Go 1.26 (`google.golang.org/grpc`, Protobuf)
- **Analytics & Compute Engine**: Rust (`rust/traffic-core`, staticlib / C-ABI FFI)
- **Frontend Core**: Vue 3 + TypeScript + Vite + Bun + Tailwind CSS + DaisyUI
- **Data Virtualization**: `@tanstack/vue-virtual`, `@tanstack/vue-table`
- **Charts & Waveforms**: Apache ECharts (`echarts`, `vue-echarts`)
- **Icons & Typography**: `@heroicons/vue`, Lucide Icons, MiSans typography

---

## 3. System Architecture

```
┌────────────────────────────────────────────────────────┐
│             sing-box Daemon (1.14+ API)                │
│    services: type: api (h2c / TLS gRPC StartedService) │
└──────────────────────────┬─────────────────────────────┘
                           │ gRPC / HTTP2 (StartedService API v4)
                           ▼
┌────────────────────────────────────────────────────────┐
│           Go gRPC Client (internal/singboxapi)         │
│  - Unary/Stream Bearer Auth Interceptor                │
│  - Auto-reconnect with Jittered Exponential Backoff    │
│  - Compatibility Probing (APIVersion == 4)             │
└──────────────────────────┬─────────────────────────────┘
                           │ Normalized Flow Events
                           ▼
┌────────────────────────────────────────────────────────┐
│      Authoritative Connection Store (internal/store)    │
│  - State Machine (NEW, UPDATE delta, CLOSED, RESET)    │
│  - TUN Traffic Filtering & Process Identification      │
│  - Ring Buffer for 50,000+ Historical Flows            │
│  - Real-Time Bandwidth & Rate Calculation              │
└──────────────────────────┬─────────────────────────────┘
                           │ Coarse-Grained Batch Snapshot
                           ▼
┌────────────────────────────────────────────────────────┐
│          Rust Analytics Core (rust/traffic-core)       │
│  - Batch Aggregation (Process, Domain, Dest, Routing)  │
│  - Top-N Ranking Computation                           │
│  - Time-Series Downsampling                            │
│  - Coarse-Grained C FFI / Pure-Go Fallback             │
└──────────────────────────┬─────────────────────────────┘
                           │ Wails v3 Direct Bindings
                           ▼
┌────────────────────────────────────────────────────────┐
│             Zashboard Style Desktop UI                 │
│  - Surge-style Live Bandwidth Waveforms & Memory Stats │
│  - Outbound Proxy Groups with Latency Badges           │
│  - Dual-mode Connection Virtual Table / Card List      │
│  - Matched Routing Rules Breakdown                     │
└────────────────────────────────────────────────────────┘
```

---

## 4. Quick Start (Windows Portable)

1. Ensure your sing-box instance (v1.14.0+) is running with the API service enabled:
   ```json
   {
     "services": [
       {
         "type": "api",
         "tag": "api",
         "listen": "127.0.0.1",
         "listen_port": 9090,
         "secret": "my-secret-token"
       }
     ]
   }
   ```
2. Download `dist/TrafficAnalyzer-windows-amd64-portable.zip` and extract `TrafficAnalyzer.exe`.
3. Double-click `TrafficAnalyzer.exe`. The application opens directly in its native desktop window.
4. Route network traffic through your sing-box TUN inbound to view live bandwidth, process attributions, and connection statistics.

---

## 5. Building from Source

```bash
# Clone the repository
git clone https://github.com/Star-Trails/sing-scope.git
cd sing-scope

# Build frontend and compile desktop application in one step
./scripts/build.sh
```

---

## 6. Testing & Benchmarks

```bash
# Run all Go unit tests:
go test -v ./internal/...

# Run high-flow synthetic benchmarks (1k, 10k, 100k flows):
go test -bench=. -benchmem ./internal/store/...

# Run Rust analytics tests:
cd rust/traffic-core && cargo test
```

---

## 7. Acknowledgements & Attribution

Special thanks and full credit to:
- **[Zephyruso/zashboard](https://github.com/Zephyruso/zashboard)** by **Zephyruso**: The UI design, component structures, layout paradigms, and theme styling in this project are based on and ported from Zashboard.
- **[SagerNet/sing-box](https://github.com/SagerNet/sing-box)**: The universal proxy platform and new 1.14+ gRPC API specifications.
- **[Wails](https://github.com/wailsapp/wails)**: The lightweight cross-platform desktop application framework.

See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for full licensing and attribution details.
