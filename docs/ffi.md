# Rust Analytics Core & FFI Architecture

## 1. Design Principles

In alignment with project architectural guidelines:
1. **Coarse-Grained Interface**: FFI is called on bulk batches (e.g. snapshots of all active and historical flows), never per-packet or per-event.
2. **Memory Safety & Ownership**: Memory allocated in Rust is freed in Rust via explicit `traffic_core_free_string`. Memory allocated in Go is freed in Go.
3. **Pure & Deterministic**: Aggregations are pure functions with no global mutable state.
4. **Zero-Dependency Portable Fallback**: When CGO is disabled (e.g. `CGO_ENABLED=0` for cross-platform portable builds), `internal/analytics` seamlessly switches to a pure-Go high-performance analytics implementation.

---

## 2. C-ABI Specification (`rust/traffic-core/src/ffi.rs`)

### 2.1 C Signatures
```c
// Return core engine version string (e.g. "0.1.0")
const char* traffic_core_version();

// Free a string buffer returned by Rust
void traffic_core_free_string(char* ptr);

// Execute batch multi-dimensional aggregation on JSON payload
char* traffic_core_analyze_batch(const char* input_json);

// Downsample high-frequency time series points into bucket averages
char* traffic_core_downsample_timeseries(const char* input_json, size_t target_buckets);
```

### 2.2 Data Exchange
- **Input Payload**: JSON string containing `flows` array, `topN` integer, and optional `inboundFilter` string.
- **Output Payload**: JSON string containing:
  - `totalFlows`, `activeFlows`
  - `totalUploadBytes`, `totalDownloadBytes`
  - `totalUploadRate`, `totalDownloadRate`
  - `byProcess` (aggregated process list with Top 5 domains and destinations per process)
  - `byDomain`, `byDestination`, `byOutbound`, `byRule`, `byProtocol`
  - `computeTimeUs` (microsecond execution timer)
