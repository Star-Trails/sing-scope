# Testing and Benchmark Results

## 1. Test Matrix Summary

| Test Suite | Package / Crate | Scope | Status |
|---|---|---|---|
| Rust Unit & Aggregation Tests | `rust/traffic-core` | Batch aggregation, inbound filter, time series downsampler | **PASS** |
| Protobuf Normalization Tests | `internal/singboxapi` | Process name extraction (Windows/Unix), `Connection`, `ConnectionEvent` conversion | **PASS** |
| Reconnect & Backoff Tests | `internal/singboxapi` | Jittered exponential backoff bounds, API compatibility check | **PASS** |
| Store State Machine Tests | `internal/store` | `NEW`, `UPDATE` delta calculation, `CLOSED` transition, `reset: true` state re-anchoring | **PASS** |
| TUN Inbound Filter Tests | `internal/store`, `internal/traffic` | Global, TUN-only, tag-specific filtering | **PASS** |
| Synthetic High-Load Test | `internal/store` | 100,000 historical flows insertion, memory, and full table query | **PASS** |
| Frontend Typecheck & Build | `frontend` | TypeScript compilation (`vue-tsc`), Vite production asset bundling | **PASS** |

---

## 2. Synthetic Benchmark Measurements

Executed on: `Intel(R) Core(TM) Ultra 9 275HX (24 vCPUs)`

```
BenchmarkConnectionStore_ProcessBatch_1000Flows-24    29667    41.4 µs/op      79 B/op    0 allocs/op
BenchmarkConnectionStore_ProcessBatch_10000Flows-24    2335   482.1 µs/op     438 B/op    0 allocs/op
BenchmarkConnectionStore_Query_10000Active-24           435     2.7 ms/op  256320 B/op 8901 allocs/op
BenchmarkConnectionStore_Overview_10000Active-24       1489   771.6 µs/op   82968 B/op  403 allocs/op
```

### 2.1 100,000 Historical Flow Stress Test
- **Ingestion Time**: 100,000 historical flows parsed and indexed in `~45 ms`.
- **Search & Sort Latency**: Querying and sorting across 100,000 historical flows completed in `~14 ms`.
