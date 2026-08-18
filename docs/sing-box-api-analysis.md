# sing-box 1.14+ New gRPC API Analysis

## 1. Upstream Source Information

- **Repository**: `https://github.com/SagerNet/sing-box`
- **Branch**: `testing`
- **Commit SHA**: `c82b9b8dc92e1495968a1e0835644e4ad6fc303b`
- **Commit Date**: `2026-08-17 17:07:47 +0800`
- **sing-box Version**: `v1.14.0-beta.17`
- **Protobuf API Version (`StartedService.GetVersion`)**: `4` (`APIVersion = 4`)

---

## 2. Protobuf Specifications & Service Definitions

Primary protobuf file: `daemon/started_service.proto` (package `daemon`, `go_package = "github.com/sagernet/sing-box/daemon"`).

### 2.1 StartedService RPC Methods

| RPC Method | Request Type | Response Type | Semantics |
|---|---|---|---|
| `GetVersion` | `google.protobuf.Empty` | `Version` | Unary. Returns `{ version: string, apiVersion: int32 }`. |
| `SubscribeServiceStatus` | `google.protobuf.Empty` | `stream ServiceStatus` | Server-streaming. Emits lifecycle status (`IDLE`, `STARTING`, `STARTED`, `STOPPING`, `FATAL`). |
| `SubscribeLog` | `google.protobuf.Empty` | `stream Log` | Server-streaming. Emits log message batches with optional `reset: true`. |
| `GetDefaultLogLevel` | `google.protobuf.Empty` | `DefaultLogLevel` | Unary. Returns default log level. |
| `ClearLogs` | `google.protobuf.Empty` | `google.protobuf.Empty` | Unary. Clears ring-buffer log memory. |
| `SubscribeStatus` | `SubscribeStatusRequest` | `stream Status` | Server-streaming. Emits memory, goroutines, connection counts, and uplink/downlink totals. |
| `SubscribeGroups` | `google.protobuf.Empty` | `stream Groups` | Server-streaming. Emits outbound groups and selector states. |
| `GetClashModeStatus` | `google.protobuf.Empty` | `ClashModeStatus` | Unary. Returns available and current Clash modes. |
| `SubscribeClashMode` | `google.protobuf.Empty` | `stream ClashMode` | Server-streaming. Emits Clash mode changes. |
| `SetClashMode` | `ClashMode` | `google.protobuf.Empty` | Unary. Switches Clash mode. |
| `URLTest` | `URLTestRequest` | `google.protobuf.Empty` | Unary. Triggers URL latency test for an outbound group. |
| `SelectOutbound` | `SelectOutboundRequest` | `google.protobuf.Empty` | Unary. Selects an outbound item in a group. |
| `SetGroupExpand` | `SetGroupExpandRequest` | `google.protobuf.Empty` | Unary. Sets UI expansion state of a group. |
| `SubscribeConnections` | `SubscribeConnectionsRequest` | `stream ConnectionEvents` | Server-streaming. Authoritative connection event stream (NEW, UPDATE, CLOSED). |
| `CloseConnection` | `CloseConnectionRequest` | `google.protobuf.Empty` | Unary. Closes a connection by its UUID string ID. |
| `CloseAllConnections` | `google.protobuf.Empty` | `google.protobuf.Empty` | Unary. Closes all active connections. |
| `GetDeprecatedWarnings` | `google.protobuf.Empty` | `DeprecatedWarnings` | Unary. Returns active configuration deprecation warnings. |
| `GetStartedAt` | `google.protobuf.Empty` | `StartedAt` | Unary. Returns instance start timestamp (Unix ms). |
| `SubscribeOutbounds` | `google.protobuf.Empty` | `stream OutboundList` | Server-streaming. Emits full list of configured outbounds. |
| `StartNetworkQualityTest` | `NetworkQualityTestRequest` | `stream NetworkQualityTestProgress` | Server-streaming. Runs Apple responsiveness / RPM test. |
| `StartSTUNTest` | `STUNTestRequest` | `stream STUNTestProgress` | Server-streaming. Tests NAT mapping & filtering behavior. |
| `SubscribeTailscaleStatus` | `google.protobuf.Empty` | `stream TailscaleStatusUpdate` | Server-streaming. Emits Tailscale endpoint status updates. |
| `SubscribeNotifications` | `google.protobuf.Empty` | `stream NotificationEvent` | Server-streaming. Emits system notifications. |

---

## 3. Connection Stream Semantics (`SubscribeConnections`)

### 3.1 Request Contract
```protobuf
message SubscribeConnectionsRequest {
  int64 interval = 1;
}
```
- **Interval Unit**: `int64` representing Go `time.Duration` in **nanoseconds**.
  - 1 second = `1_000_000_000` (i.e. `int64(time.Second)`).
  - If `interval <= 0`, server defaults to `1 * time.Second`.

### 3.2 Response Contract & Event Lifecycle
```protobuf
enum ConnectionEventType {
  CONNECTION_EVENT_NEW = 0;
  CONNECTION_EVENT_UPDATE = 1;
  CONNECTION_EVENT_CLOSED = 2;
}

message ConnectionEvent {
  ConnectionEventType type = 1;
  string id = 2;
  Connection connection = 3;
  int64 uplinkDelta = 4;
  int64 downlinkDelta = 5;
  int64 closedAt = 6;
}

message ConnectionEvents {
  repeated ConnectionEvent events = 1;
  bool reset = 2;
}
```

### 3.3 Server Event Dispatch Details
1. **Initial Snapshot (`reset: true`)**:
   - Immediately upon stream establishment, the server sends a `ConnectionEvents` batch with `reset: true`.
   - Contains all currently active connections as `CONNECTION_EVENT_NEW` with full `Connection` metadata.
   - Also contains recent closed connections in ring-buffer as `CONNECTION_EVENT_NEW` (with `closedAt` populated).
   - **Client Requirement**: Client MUST clear local active flow table and rebuild from this batch.
2. **Real-Time Lifecycle Events**:
   - When a new connection is established, an immediate event `CONNECTION_EVENT_NEW` is pushed containing full `Connection` metadata.
   - When a connection closes, an immediate event `CONNECTION_EVENT_CLOSED` is pushed with `closedAt` (Unix ms) and final `Connection` metadata snapshot.
3. **Periodic Traffic Updates (`CONNECTION_EVENT_UPDATE`)**:
   - Fired by ticker every `interval`.
   - Computes `uplinkDelta = currentUpload - lastUploadSnapshot` and `downlinkDelta = currentDownload - lastDownloadSnapshot`.
   - If `uplinkDelta > 0 || downlinkDelta > 0`, emits `CONNECTION_EVENT_UPDATE` containing `id`, `uplinkDelta`, and `downlinkDelta`.
   - If counters reset (e.g. overflow or tracker rebuild), `uplinkDelta` and `downlinkDelta` are zeroed and snapshots re-anchored.

### 3.4 Rate Calculation Invariant
```
uploadRate   = uplinkDelta   / elapsed_seconds
downloadRate = downlinkDelta / elapsed_seconds
```
Client records the exact timestamp of each update to calculate `elapsed_seconds = (now - lastUpdateAt).Seconds()`, ensuring accurate rates even under system scheduling variations.

---

## 4. Connection Metadata Schema

```protobuf
message Connection {
  string id = 1;
  string inbound = 2;
  string inboundType = 3;
  int32 ipVersion = 4;
  string network = 5;
  string source = 6;
  string destination = 7;
  string domain = 8;
  string protocol = 9;
  string user = 10;
  string fromOutbound = 11;
  int64 createdAt = 12;
  int64 closedAt = 13;
  int64 uplink = 14;
  int64 downlink = 15;
  int64 uplinkTotal = 16;
  int64 downlinkTotal = 17;
  string rule = 18;
  string outbound = 19;
  string outboundType = 20;
  repeated string chainList = 21;
  ProcessInfo processInfo = 22;
}

message ProcessInfo {
  uint32 processId = 1;
  int32 userId = 2;
  string userName = 3;
  string processPath = 4;
  repeated string packageNames = 5;
}
```

### 4.1 TUN Inbound Identification
- `inboundType`: Set to `"tun"` for TUN devices (e.g. `"tun"`).
- `inbound`: User-defined inbound tag (e.g. `"tun-in"`, `"main"`, `"vpn"`).
- **Filtering Rule**:
  - Filter by `inboundType == "tun"` for TUN traffic.
  - Optionally filter by `inbound == selectedTag` for multi-TUN configs.

---

## 5. Global Status Stream (`SubscribeStatus`)

```protobuf
message Status {
  uint64 memory = 1;
  int32 goroutines = 2;
  int32 connectionsIn = 3;
  int32 connectionsOut = 4;
  bool trafficAvailable = 5;
  int64 uplink = 6;
  int64 downlink = 7;
  int64 uplinkTotal = 8;
  int64 downlinkTotal = 9;
}
```
- Emits sing-box process-level resource metrics (Go runtime heap memory, goroutine count, overall in/out connection counters, global uplink/downlink rates and totals).

---

## 6. Authentication & Network Transport

### 6.1 Authentication Mechanism
- Header format: `authorization: Bearer <secret>` in gRPC metadata / HTTP/2 request headers.
- If `secret` is configured on sing-box API service, any request lacking a matching bearer token returns gRPC status code `codes.Unauthenticated` (`16`).
- If `secret` is empty on server, authentication is bypassed.

### 6.2 Transport (h2c vs TLS)
- Sing-box API service uses `h2c.NewHandler(..., &http2.Server{})`.
- **`http://host:port`**: Communicates via standard HTTP/2 cleartext (h2c) using `grpc.WithTransportCredentials(insecure.NewCredentials())`.
- **`https://host:port`**: Communicates via standard HTTP/2 over TLS using `grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{ServerName: host}))`.
- The Go gRPC client automatically chooses `insecure.NewCredentials()` or TLS credentials based on URL scheme.

---

## 7. Reconnection & Resilience Strategy

1. **State Transitions**:
   `Disconnected` $\rightarrow$ `Connecting` $\rightarrow$ `Connected` $\rightarrow$ `Reconnecting` $\rightarrow$ `Connected`
2. **Backoff**: Bounded exponential backoff (initial: 500ms, max: 8s, multiplier: 1.5, jitter: $\pm 20\%$).
3. **Verification on Reconnect**:
   - Dial target and execute `GetVersion(ctx)`.
   - Verify `apiVersion >= 4`.
   - Start `SubscribeConnections(ctx, &SubscribeConnectionsRequest{Interval: 1_000_000_000})`.
   - Receive initial snapshot with `reset: true`.
   - Re-anchor local authoritative state.
4. **Cancellation**: Context cancellation cleanly shuts down all background goroutines, streams, and active gRPC channels.

---

## 8. Compatibility Policy

- API Version check: `apiVersion == 4` is fully supported.
- If `apiVersion < 4` or `apiVersion > 4`, display a structured error warning with version details while attempting best-effort stream processing where schemas align.
