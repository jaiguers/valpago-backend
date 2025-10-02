use axum::{Router, routing::get};
use mongodb::Database;
use redis::Client as RedisClient;

pub mod users;
pub mod auth;
pub mod transactions;
use crate::realtime::ws;

use crate::realtime::notifier::Notifier;

#[derive(Clone)]
pub struct AppState {
    pub mongodb: Database,
    pub redis: RedisClient,
    pub notifier: Notifier,
}

impl AppState {
    pub fn new(mongodb: Database, redis: RedisClient, notifier: Notifier) -> Self {
        Self { mongodb, redis, notifier }
    }
}

pub fn health_router() -> Router<AppState> {
    Router::new().route("/health", get(|| async { "ok" }))
}

pub fn api_router() -> Router<AppState> {
    Router::new()
        .nest("/users", users::router())
        .nest("/auth", auth::router())
        .nest("/transactions", transactions::router())
        .merge(ws::router())
}

