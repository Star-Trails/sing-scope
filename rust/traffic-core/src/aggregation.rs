use std::collections::HashMap;
use std::time::Instant;

use crate::models::*;

pub fn matches_inbound_filter(flow: &FlowInput, filter: Option<&str>) -> bool {
    match filter {
        None | Some("") | Some("all") => true,
        Some("tun") | Some("tun:all") => flow.inbound_type.eq_ignore_ascii_case("tun"),
        Some(tag) if tag.starts_with("tag:") => {
            let actual = &tag[4..];
            flow.inbound == actual
        }
        Some(val) => flow.inbound == val || flow.inbound_type.eq_ignore_ascii_case(val),
    }
}

pub fn analyze_batch(req: BatchAnalysisRequest) -> BatchAnalysisResult {
    let start_time = Instant::now();
    let filter = req.inbound_filter.as_deref();

    let mut total_flows = 0usize;
    let mut active_flows = 0usize;
    let mut total_upload_bytes = 0i64;
    let mut total_download_bytes = 0i64;
    let mut total_upload_rate = 0.0f64;
    let mut total_download_rate = 0.0f64;

    // Aggregation maps
    let mut domain_map: HashMap<String, NamedAggregate> = HashMap::new();
    let mut dest_map: HashMap<String, NamedAggregate> = HashMap::new();
    let mut outbound_map: HashMap<String, NamedAggregate> = HashMap::new();
    let mut rule_map: HashMap<String, NamedAggregate> = HashMap::new();
    let mut protocol_map: HashMap<String, NamedAggregate> = HashMap::new();

    // Process tracking
    struct ProcessAcc {
        name: String,
        path: String,
        pid: u32,
        conn_count: usize,
        active_count: usize,
        upload_total: i64,
        download_total: i64,
        upload_rate: f64,
        download_rate: f64,
        domains: HashMap<String, i64>,
        destinations: HashMap<String, i64>,
    }
    let mut process_map: HashMap<String, ProcessAcc> = HashMap::new();

    for flow in &req.flows {
        if !matches_inbound_filter(flow, filter) {
            continue;
        }

        total_flows += 1;
        if flow.is_active {
            active_flows += 1;
            total_upload_rate += flow.upload_rate;
            total_download_rate += flow.download_rate;
        }
        total_upload_bytes += flow.upload_total;
        total_download_bytes += flow.download_total;

        let total_bytes = flow.upload_total + flow.download_total;
        let total_rate = flow.upload_rate + flow.download_rate;

        // Domain
        if !flow.domain.is_empty() {
            let entry = domain_map
                .entry(flow.domain.clone())
                .or_insert_with(|| NamedAggregate {
                    key: flow.domain.clone(),
                    name: flow.domain.clone(),
                    category: None,
                    connection_count: 0,
                    active_count: 0,
                    upload_total: 0,
                    download_total: 0,
                    total_bytes: 0,
                    upload_rate: 0.0,
                    download_rate: 0.0,
                    total_rate: 0.0,
                });
            entry.connection_count += 1;
            if flow.is_active {
                entry.active_count += 1;
                entry.upload_rate += flow.upload_rate;
                entry.download_rate += flow.download_rate;
                entry.total_rate += total_rate;
            }
            entry.upload_total += flow.upload_total;
            entry.download_total += flow.download_total;
            entry.total_bytes += total_bytes;
        }

        // Destination
        if !flow.destination.is_empty() {
            let entry = dest_map
                .entry(flow.destination.clone())
                .or_insert_with(|| NamedAggregate {
                    key: flow.destination.clone(),
                    name: flow.destination.clone(),
                    category: None,
                    connection_count: 0,
                    active_count: 0,
                    upload_total: 0,
                    download_total: 0,
                    total_bytes: 0,
                    upload_rate: 0.0,
                    download_rate: 0.0,
                    total_rate: 0.0,
                });
            entry.connection_count += 1;
            if flow.is_active {
                entry.active_count += 1;
                entry.upload_rate += flow.upload_rate;
                entry.download_rate += flow.download_rate;
                entry.total_rate += total_rate;
            }
            entry.upload_total += flow.upload_total;
            entry.download_total += flow.download_total;
            entry.total_bytes += total_bytes;
        }

        // Outbound
        if !flow.outbound.is_empty() {
            let entry = outbound_map
                .entry(flow.outbound.clone())
                .or_insert_with(|| NamedAggregate {
                    key: flow.outbound.clone(),
                    name: flow.outbound.clone(),
                    category: if flow.outbound_type.is_empty() {
                        None
                    } else {
                        Some(flow.outbound_type.clone())
                    },
                    connection_count: 0,
                    active_count: 0,
                    upload_total: 0,
                    download_total: 0,
                    total_bytes: 0,
                    upload_rate: 0.0,
                    download_rate: 0.0,
                    total_rate: 0.0,
                });
            entry.connection_count += 1;
            if flow.is_active {
                entry.active_count += 1;
                entry.upload_rate += flow.upload_rate;
                entry.download_rate += flow.download_rate;
                entry.total_rate += total_rate;
            }
            entry.upload_total += flow.upload_total;
            entry.download_total += flow.download_total;
            entry.total_bytes += total_bytes;
        }

        // Rule
        if !flow.rule.is_empty() {
            let entry = rule_map
                .entry(flow.rule.clone())
                .or_insert_with(|| NamedAggregate {
                    key: flow.rule.clone(),
                    name: flow.rule.clone(),
                    category: None,
                    connection_count: 0,
                    active_count: 0,
                    upload_total: 0,
                    download_total: 0,
                    total_bytes: 0,
                    upload_rate: 0.0,
                    download_rate: 0.0,
                    total_rate: 0.0,
                });
            entry.connection_count += 1;
            if flow.is_active {
                entry.active_count += 1;
                entry.upload_rate += flow.upload_rate;
                entry.download_rate += flow.download_rate;
                entry.total_rate += total_rate;
            }
            entry.upload_total += flow.upload_total;
            entry.download_total += flow.download_total;
            entry.total_bytes += total_bytes;
        }

        // Protocol
        if !flow.protocol.is_empty() {
            let entry = protocol_map
                .entry(flow.protocol.clone())
                .or_insert_with(|| NamedAggregate {
                    key: flow.protocol.clone(),
                    name: flow.protocol.clone(),
                    category: None,
                    connection_count: 0,
                    active_count: 0,
                    upload_total: 0,
                    download_total: 0,
                    total_bytes: 0,
                    upload_rate: 0.0,
                    download_rate: 0.0,
                    total_rate: 0.0,
                });
            entry.connection_count += 1;
            if flow.is_active {
                entry.active_count += 1;
                entry.upload_rate += flow.upload_rate;
                entry.download_rate += flow.download_rate;
                entry.total_rate += total_rate;
            }
            entry.upload_total += flow.upload_total;
            entry.download_total += flow.download_total;
            entry.total_bytes += total_bytes;
        }

        // Process
        let (pname, ppath, pid) = match &flow.process {
            Some(p) => {
                let mut name = if !p.process_name.is_empty() && p.process_name != "Unknown" {
                    p.process_name.clone()
                } else if !p.process_path.is_empty() {
                    let normalized = p.process_path.replace('\\', "/");
                    let clean = normalized.trim_start_matches(':');
                    let base = clean.split('/').last().unwrap_or(clean);
                    if base.is_empty() { String::new() } else { base.to_string() }
                } else if let Some(first_pkg) = p.package_names.as_ref().and_then(|pkgs| pkgs.first()) {
                    first_pkg.clone()
                } else if p.process_id > 0 {
                    format!("PID:{}", p.process_id)
                } else {
                    String::new()
                };
                if name.is_empty() {
                    name = "Unknown".to_string();
                }
                (name, p.process_path.clone(), p.process_id)
            }
            None => ("Unknown".to_string(), String::new(), 0),
        };

        let proc_entry = process_map.entry(pname.clone()).or_insert_with(|| ProcessAcc {
            name: pname,
            path: ppath,
            pid,
            conn_count: 0,
            active_count: 0,
            upload_total: 0,
            download_total: 0,
            upload_rate: 0.0,
            download_rate: 0.0,
            domains: HashMap::new(),
            destinations: HashMap::new(),
        });
        proc_entry.conn_count += 1;
        if flow.is_active {
            proc_entry.active_count += 1;
            proc_entry.upload_rate += flow.upload_rate;
            proc_entry.download_rate += flow.download_rate;
        }
        proc_entry.upload_total += flow.upload_total;
        proc_entry.download_total += flow.download_total;
        if !flow.domain.is_empty() {
            *proc_entry.domains.entry(flow.domain.clone()).or_insert(0) += total_bytes;
        }
        if !flow.destination.is_empty() {
            *proc_entry.destinations.entry(flow.destination.clone()).or_insert(0) += total_bytes;
        }
    }

    let top_n = req.top_n;

    // Convert & sort
    let mut by_domain = domain_map.into_values().collect::<Vec<_>>();
    by_domain.sort_by(|a, b| b.total_bytes.cmp(&a.total_bytes));
    by_domain.truncate(top_n);

    let mut by_destination = dest_map.into_values().collect::<Vec<_>>();
    by_destination.sort_by(|a, b| b.total_bytes.cmp(&a.total_bytes));
    by_destination.truncate(top_n);

    let mut by_outbound = outbound_map.into_values().collect::<Vec<_>>();
    by_outbound.sort_by(|a, b| b.total_bytes.cmp(&a.total_bytes));
    by_outbound.truncate(top_n);

    let mut by_rule = rule_map.into_values().collect::<Vec<_>>();
    by_rule.sort_by(|a, b| b.total_bytes.cmp(&a.total_bytes));
    by_rule.truncate(top_n);

    let mut by_protocol = protocol_map.into_values().collect::<Vec<_>>();
    by_protocol.sort_by(|a, b| b.total_bytes.cmp(&a.total_bytes));
    by_protocol.truncate(top_n);

    let mut by_process = Vec::with_capacity(process_map.len());
    for (_, p) in process_map {
        let total_b = p.upload_total + p.download_total;

        let mut top_doms: Vec<NamedAggregate> = p
            .domains
            .into_iter()
            .map(|(k, bytes)| NamedAggregate {
                key: k.clone(),
                name: k,
                category: None,
                connection_count: 0,
                active_count: 0,
                upload_total: 0,
                download_total: 0,
                total_bytes: bytes,
                upload_rate: 0.0,
                download_rate: 0.0,
                total_rate: 0.0,
            })
            .collect();
        top_doms.sort_by(|a, b| b.total_bytes.cmp(&a.total_bytes));
        top_doms.truncate(5);

        let mut top_dests: Vec<NamedAggregate> = p
            .destinations
            .into_iter()
            .map(|(k, bytes)| NamedAggregate {
                key: k.clone(),
                name: k,
                category: None,
                connection_count: 0,
                active_count: 0,
                upload_total: 0,
                download_total: 0,
                total_bytes: bytes,
                upload_rate: 0.0,
                download_rate: 0.0,
                total_rate: 0.0,
            })
            .collect();
        top_dests.sort_by(|a, b| b.total_bytes.cmp(&a.total_bytes));
        top_dests.truncate(5);

        by_process.push(ProcessAggregate {
            process_name: p.name,
            process_path: p.path,
            process_id: p.pid,
            connection_count: p.conn_count,
            active_count: p.active_count,
            upload_total: p.upload_total,
            download_total: p.download_total,
            total_bytes: total_b,
            upload_rate: p.upload_rate,
            download_rate: p.download_rate,
            top_domains: top_doms,
            top_destinations: top_dests,
        });
    }
    by_process.sort_by(|a, b| b.total_bytes.cmp(&a.total_bytes));
    by_process.truncate(top_n);

    let compute_time_us = start_time.elapsed().as_micros() as u64;

    BatchAnalysisResult {
        total_flows,
        active_flows,
        total_upload_bytes,
        total_download_bytes,
        total_upload_rate,
        total_download_rate,
        by_process,
        by_domain,
        by_destination,
        by_outbound,
        by_rule,
        by_protocol,
        compute_time_us,
    }
}

pub fn downsample_timeseries(
    points: &[TimeSeriesPoint],
    target_buckets: usize,
) -> Vec<TimeSeriesPoint> {
    if points.is_empty() || target_buckets == 0 || points.len() <= target_buckets {
        return points.to_vec();
    }

    let bucket_size = (points.len() as f64) / (target_buckets as f64);
    let mut downsampled = Vec::with_capacity(target_buckets);

    for i in 0..target_buckets {
        let start_idx = (i as f64 * bucket_size).floor() as usize;
        let end_idx = (((i + 1) as f64 * bucket_size).floor() as usize).min(points.len());

        if start_idx >= end_idx {
            continue;
        }

        let slice = &points[start_idx..end_idx];
        let count = slice.len() as f64;
        let mut sum_up = 0.0f64;
        let mut sum_down = 0.0f64;
        let mut sum_ts = 0i64;
        let mut max_active = 0usize;

        for p in slice {
            sum_up += p.upload_rate;
            sum_down += p.download_rate;
            sum_ts += p.timestamp;
            if p.active_flows > max_active {
                max_active = p.active_flows;
            }
        }

        downsampled.push(TimeSeriesPoint {
            timestamp: sum_ts / (slice.len() as i64),
            upload_rate: sum_up / count,
            download_rate: sum_down / count,
            active_flows: max_active,
        });
    }

    downsampled
}
