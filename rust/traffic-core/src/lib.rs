pub mod aggregation;
pub mod ffi;
pub mod models;

pub use aggregation::{analyze_batch, downsample_timeseries, matches_inbound_filter};
pub use ffi::*;
pub use models::*;

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_batch_analysis_and_filter() {
        let req = BatchAnalysisRequest {
            flows: vec![
                FlowInput {
                    id: "1".to_string(),
                    inbound: "tun-in".to_string(),
                    inbound_type: "tun".to_string(),
                    ip_version: 4,
                    network: "tcp".to_string(),
                    source: "192.168.1.5:1234".to_string(),
                    destination: "1.1.1.1:443".to_string(),
                    domain: "cloudflare.com".to_string(),
                    protocol: "tls".to_string(),
                    user: "".to_string(),
                    from_outbound: "".to_string(),
                    rule: "direct".to_string(),
                    outbound: "direct".to_string(),
                    outbound_type: "direct".to_string(),
                    chain_list: vec![],
                    upload_total: 1000,
                    download_total: 5000,
                    upload_rate: 100.0,
                    download_rate: 500.0,
                    is_active: true,
                    created_at: None,
                    closed_at: None,
                },
                FlowInput {
                    id: "2".to_string(),
                    inbound: "mixed-in".to_string(),
                    inbound_type: "mixed".to_string(),
                    ip_version: 4,
                    network: "udp".to_string(),
                    source: "192.168.1.5:1235".to_string(),
                    destination: "8.8.8.8:53".to_string(),
                    domain: "dns.google".to_string(),
                    protocol: "dns".to_string(),
                    user: "".to_string(),
                    from_outbound: "".to_string(),
                    rule: "dns-rule".to_string(),
                    outbound: "proxy".to_string(),
                    outbound_type: "vless".to_string(),
                    chain_list: vec![],
                    upload_total: 200,
                    download_total: 400,
                    upload_rate: 0.0,
                    download_rate: 0.0,
                    is_active: false,
                    created_at: None,
                    closed_at: None,
                },
            ],
            top_n: 10,
            inbound_filter: Some("tun".to_string()),
        };

        let result = analyze_batch(req);
        assert_eq!(result.total_flows, 1);
        assert_eq!(result.active_flows, 1);
        assert_eq!(result.total_upload_bytes, 1000);
        assert_eq!(result.total_download_bytes, 5000);
        assert_eq!(result.by_domain.len(), 1);
        assert_eq!(result.by_domain[0].name, "cloudflare.com");
    }

    #[test]
    fn test_downsample_timeseries() {
        let points = (0..100)
            .map(|i| TimeSeriesPoint {
                timestamp: i * 1000,
                upload_rate: (i * 10) as f64,
                download_rate: (i * 20) as f64,
                active_flows: (i % 5) as usize,
            })
            .collect::<Vec<_>>();

        let downsampled = downsample_timeseries(&points, 10);
        assert_eq!(downsampled.len(), 10);
        assert!(downsampled[0].upload_rate < downsampled[9].upload_rate);
    }
}
