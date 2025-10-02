use axum::{routing::post, Json, Router};
use axum::extract::State;
use serde::{Deserialize, Serialize};
use mongodb::bson::doc;

use crate::{errors::{ApiError, ApiResult}, routes::AppState, security::{jwt, password}};

#[derive(Debug, Deserialize)]
pub struct LoginDto {
    pub email: String,
    pub password: String,
}

#[derive(Debug, Serialize)]
pub struct LoginResponse {
    pub token: String,
}

pub fn router() -> Router<AppState> {
    Router::new().route("/login", post(login))
}

async fn login(State(state): State<AppState>, Json(payload): Json<LoginDto>) -> ApiResult<Json<LoginResponse>> {
    let coll = state.mongodb.collection::<crate::routes::users::User>("users");
    let found = coll
        .find_one(doc!{"email": &payload.email}, None)
        .await
        .map_err(|e| ApiError::Internal(e.to_string()))?;
    let user = found.ok_or_else(|| ApiError::Unauthorized("invalid credentials".into()))?;
    if !password::verify_password(&user.password, &payload.password) {
        return Err(ApiError::Unauthorized("invalid credentials".into()));
    }
    let token = jwt::create_jwt(&user.id.unwrap().to_hex()).map_err(|e| ApiError::Internal(e.to_string()))?;
    Ok(Json(LoginResponse { token }))
}

