use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct FlowInput {
    pub id: String,
    #[serde(default)]
    pub inbound: String,
    #[serde(default)]
    pub inbound_type: String,
    #[serde(default)]
    pub ip_version: i32,
    #[serde(default)]
    pub network: String,
    #[serde(default)]
    pub source: String,
    #[serde(default)]
    pub destination: String,
    #[serde(default)]
    pub domain: String,
    #[serde(default)]
    pub protocol: String,
    #[serde(default)]
    pub user: String,
    #[serde(default)]
    pub from_outbound: String,
    #[serde(default)]
    pub rule: String,
    #[serde(default)]
    pub outbound: String,
    #[serde(default)]
    pub outbound_type: String,
    #[serde(default)]
    pub chain_list: Vec<String>,
    #[serde(default)]
    pub upload_total: i64,
    #[serde(default)]
    pub download_total: i64,
    #[serde(default)]
    pub upload_rate: f64,
    #[serde(default)]
    pub download_rate: f64,
    #[serde(default)]
    pub is_active: bool,
    #[serde(default)]
    pub created_at: Option<String>,
    #[serde(default)]
    pub closed_at: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct NamedAggregate {
    pub key: String,
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub category: Option<String>,
    pub connection_count: usize,
    pub active_count: usize,
    pub upload_total: i64,
    pub download_total: i64,
    pub total_bytes: i64,
    pub upload_rate: f64,
    pub download_rate: f64,
    pub total_rate: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct BatchAnalysisRequest {
    pub flows: Vec<FlowInput>,
    #[serde(default = "default_top_n")]
    pub top_n: usize,
    #[serde(default)]
    pub inbound_filter: Option<String>,
}

fn default_top_n() -> usize {
    10
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct BatchAnalysisResult {
    pub total_flows: usize,
    pub active_flows: usize,
    pub total_upload_bytes: i64,
    pub total_download_bytes: i64,
    pub total_upload_rate: f64,
    pub total_download_rate: f64,
    pub by_domain: Vec<NamedAggregate>,
    pub by_destination: Vec<NamedAggregate>,
    pub by_outbound: Vec<NamedAggregate>,
    pub by_rule: Vec<NamedAggregate>,
    pub by_protocol: Vec<NamedAggregate>,
    pub compute_time_us: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct TimeSeriesPoint {
    pub timestamp: i64,
    pub upload_rate: f64,
    pub download_rate: f64,
    #[serde(default)]
    pub active_flows: usize,
}
