use axum::{extract::ws::{WebSocketUpgrade, WebSocket, Message}, response::IntoResponse, extract::State, routing::get, Router};
use futures::{StreamExt};

use crate::routes::AppState;

pub fn router() -> Router<AppState> {
    Router::new().route("/ws", get(ws_handler))
}

async fn ws_handler(State(state): State<AppState>, ws: WebSocketUpgrade) -> impl IntoResponse {
    ws.on_upgrade(move |socket| handle_socket(state, socket))
}

async fn handle_socket(state: AppState, mut socket: WebSocket) {
    let mut rx = state.notifier.subscribe();
    // Forward notifications to this socket
    tokio::spawn(async move {
        while let Ok(msg) = rx.recv().await {
            if socket.send(Message::Text(msg)).await.is_err() { break; }
        }
    });
}

