use std::net::SocketAddr;

use axum::{Router};
use tower_http::{cors::CorsLayer, trace::TraceLayer};
use tracing_subscriber::EnvFilter;

mod config;
mod errors;
mod db;
mod security;
mod routes;
mod realtime;
mod worker;

#[tokio::main]
async fn main() {
    // Tracing setup
    let env_filter = std::env::var("RUST_LOG").unwrap_or_else(|_| "info,tower_http=debug".into());
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::new(env_filter))
        .init();

    // Load env
    config::load_dotenv();

    // Build shared state (db, redis, broadcaster)
    let mongodb = db::mongodb::create_mongo_client().await.expect("mongo client");
    let redis = db::redis::create_redis_client().expect("redis client");
    let notifier = realtime::notifier::create_notifier();

    let app_state = routes::AppState::new(mongodb.clone(), redis.clone(), notifier.clone());

    // Build router
    let app = Router::new()
        .merge(routes::health_router())
        .nest("/api", routes::api_router())
        .with_state(app_state)
        .layer(TraceLayer::new_for_http())
        .layer(config::cors_layer());

    // Start background worker
    worker::spawn_redis_worker(mongodb, redis, notifier);

    let port = std::env::var("SERVER_PORT").unwrap_or_else(|_| "8080".into());
    let addr: SocketAddr = format!("0.0.0.0:{}", port).parse().expect("invalid port");
    tracing::info!("listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.expect("bind failed");
    axum::serve(listener, app).await.expect("server failed");
}

