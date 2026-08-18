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

    let compute_time_us = start_time.elapsed().as_micros() as u64;

    BatchAnalysisResult {
        total_flows,
        active_flows,
        total_upload_bytes,
        total_download_bytes,
        total_upload_rate,
        total_download_rate,
        by_domain,
        by_destination,
        by_outbound,
        by_rule,
        by_protocol,
        compute_time_us,
    }
}
