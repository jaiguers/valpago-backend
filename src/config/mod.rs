use std::time::Duration;

use tower_http::cors::{Any, CorsLayer};
use axum::http;

pub fn load_dotenv() {
    let _ = dotenvy::dotenv();
}

pub fn cors_layer() -> CorsLayer {
    let allowed = std::env::var("ALLOWED_ORIGINS").unwrap_or_else(|_| "*".into());
    if allowed == "*" {
        CorsLayer::very_permissive()
    } else {
        let origin: http::HeaderValue = allowed.parse().unwrap_or(http::HeaderValue::from_static("*"));
        CorsLayer::new()
            .allow_origin(origin)
            .allow_methods([http::Method::GET, http::Method::POST, http::Method::PUT])
            .allow_headers(Any)
            .max_age(Duration::from_secs(60 * 60))
    }
}

