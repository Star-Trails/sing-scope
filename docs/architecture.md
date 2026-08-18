# sing-scope Architecture

## 1. System Overview

`sing-scope` is a cross-platform desktop application designed to analyze network connections and traffic exposed by the new sing-box 1.14+ gRPC API (`StartedService`).

```
┌────────────────────────────────────────────────────────┐
│             sing-box Daemon (1.14+ API)                │
│    services: type: api (h2c / TLS gRPC Server)         │
└──────────────────────────┬─────────────────────────────┘
                           │ gRPC / HTTP2 (StartedService)
                           ▼
┌────────────────────────────────────────────────────────┐
│           Go gRPC Client (internal/singboxapi)         │
│  - Unary/Stream Bearer Auth Interceptor                │
│  - Dynamic Transport (h2c cleartext / TLS)             │
│  - Auto-reconnect with Jittered Exponential Backoff    │
│  - Compatibility Probing (APIVersion == 4)             │
└──────────────────────────┬─────────────────────────────┘
                           │ Normalized Flow Events
                           ▼
┌────────────────────────────────────────────────────────┐
│      Authoritative Connection Store (internal/store)    │
│  - State Machine (NEW, UPDATE delta, CLOSED, RESET)    │
│  - Inbound & TUN Traffic Filter Pipeline               │
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
                           │ Desktop RPC & Embedded SPA Server
                           ▼
┌────────────────────────────────────────────────────────┐
│            Vue 3 + TypeScript Desktop UI               │
│  - TanStack Virtual Table (10,000+ Flows at 60 FPS)    │
│  - Apache ECharts Real-Time Bandwidth Visualization    │
│  - Pinia State Management & Tailwind CSS Theme Engine  │
│  - Restrained, Information-Dense Desktop Design        │
└────────────────────────────────────────────────────────┘
```

## 2. Core Subsystems

### 2.1 Go gRPC Client (`internal/singboxapi`)
- Communicates with sing-box exclusively through `StartedService` protobuf definition.
- Manages connection lifecycle (`SubscribeConnections`, `SubscribeStatus`, `SubscribeLog`, `SubscribeGroups`).
- Normalizes raw protobuf structures into immutable application domain models (`domain.Flow`, `domain.ProcessInfo`, `domain.SystemStatus`).

### 2.2 Authoritative Connection Store (`internal/store`)
- Receives stream event batches.
- Handles `reset: true` by discarding stale active connection state and re-anchoring from the initial server snapshot.
- Computes upload/download rates from `uplinkDelta` and `downlinkDelta` divided by actual elapsed time.
- Manages ring-buffer history for closed connections.

### 2.3 Rust Analytics Core (`rust/traffic-core`)
- High-performance, pure and deterministic batch analytics engine.
- Exposed via coarse-grained C-ABI functions (`traffic_core_analyze_batch`, `traffic_core_downsample_timeseries`).
- Linked statically into the Go binary (`libtraffic_core.a`) with pure-Go fallback when CGO is disabled.

### 2.4 Desktop Frontend (`frontend/`)
- Built with Vue 3, Vite, Bun, Tailwind CSS, TanStack Table & Virtual, and Apache ECharts.
- Provides 7 dedicated views: Overview, Connections, Processes, Destinations, Routing, Logs, and Settings.
