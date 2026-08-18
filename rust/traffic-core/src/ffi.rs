use std::ffi::{CStr, CString};
use std::os::raw::c_char;

use crate::aggregation::{analyze_batch, downsample_timeseries};
use crate::models::{BatchAnalysisRequest, TimeSeriesPoint};

const VERSION: &str = "0.1.0\0";

#[no_mangle]
pub extern "C" fn traffic_core_version() -> *const c_char {
    VERSION.as_ptr() as *const c_char
}

#[no_mangle]
pub unsafe extern "C" fn traffic_core_free_string(ptr: *mut c_char) {
    if !ptr.is_null() {
        drop(CString::from_raw(ptr));
    }
}

#[no_mangle]
pub unsafe extern "C" fn traffic_core_analyze_batch(input_json: *const c_char) -> *mut c_char {
    if input_json.is_null() {
        return std::ptr::null_mut();
    }

    let c_str = match CStr::from_ptr(input_json).to_str() {
        Ok(s) => s,
        Err(_) => return std::ptr::null_mut(),
    };

    let req: BatchAnalysisRequest = match serde_json::from_str(c_str) {
        Ok(r) => r,
        Err(e) => {
            let err_json = format!(r#"{{"error":"{}"}}"#, e);
            return CString::new(err_json).unwrap().into_raw();
        }
    };

    let result = analyze_batch(req);
    match serde_json::to_string(&result) {
        Ok(json_str) => match CString::new(json_str) {
            Ok(c_res) => c_res.into_raw(),
            Err(_) => std::ptr::null_mut(),
        },
        Err(_) => std::ptr::null_mut(),
    }
}

#[no_mangle]
pub unsafe extern "C" fn traffic_core_downsample_timeseries(
    input_json: *const c_char,
    target_buckets: usize,
) -> *mut c_char {
    if input_json.is_null() {
        return std::ptr::null_mut();
    }

    let c_str = match CStr::from_ptr(input_json).to_str() {
        Ok(s) => s,
        Err(_) => return std::ptr::null_mut(),
    };

    let points: Vec<TimeSeriesPoint> = match serde_json::from_str(c_str) {
        Ok(p) => p,
        Err(e) => {
            let err_json = format!(r#"{{"error":"{}"}}"#, e);
            return CString::new(err_json).unwrap().into_raw();
        }
    };

    let result = downsample_timeseries(&points, target_buckets);
    match serde_json::to_string(&result) {
        Ok(json_str) => match CString::new(json_str) {
            Ok(c_res) => c_res.into_raw(),
            Err(_) => std::ptr::null_mut(),
        },
        Err(_) => std::ptr::null_mut(),
    }
}
