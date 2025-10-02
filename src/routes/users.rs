use axum::{extract::{Path, State}, routing::{get, post, put}, Json, Router};
use mongodb::{bson::{doc, oid::ObjectId}, Database};
use serde::{Deserialize, Serialize};
use futures::TryStreamExt;

use crate::{errors::{ApiError, ApiResult}, security::password, routes::AppState};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct User {
    #[serde(rename = "_id", skip_serializing_if = "Option::is_none")]
    pub id: Option<ObjectId>,
    pub name: String,
    pub lastname: String,
    pub email: String,
    pub password: String,
    pub celular: String,
}

#[derive(Debug, Deserialize)]
pub struct CreateUserDto {
    pub name: String,
    pub lastname: String,
    pub email: String,
    pub password: String,
    pub celular: String,
}

#[derive(Debug, Deserialize)]
pub struct UpdateUserDto {
    pub name: Option<String>,
    pub lastname: Option<String>,
    pub celular: Option<String>,
}

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/", post(create_user).get(list_users))
        .route("/:id", get(get_user_by_id).put(update_user))
}

async fn create_user(State(state): State<AppState>, Json(payload): Json<CreateUserDto>) -> ApiResult<Json<User>> {
    let mut user = User {
        id: None,
        name: payload.name,
        lastname: payload.lastname,
        email: payload.email,
        password: password::hash_password(&payload.password).map_err(|e| ApiError::Internal(e.to_string()))?,
        celular: payload.celular,
    };
    let coll = state.mongodb.collection::<User>("users");
    coll.insert_one(&user, None).await.map_err(|e| ApiError::Internal(e.to_string()))?;
    Ok(Json(user))
}

async fn list_users(State(state): State<AppState>) -> ApiResult<Json<Vec<User>>> {
    let coll = state.mongodb.collection::<User>("users");
    let mut cursor = coll.find(None, None).await.map_err(|e| ApiError::Internal(e.to_string()))?;
    let mut out = Vec::new();
    while let Some(doc) = cursor.try_next().await.map_err(|e| ApiError::Internal(e.to_string()))? { out.push(doc); }
    Ok(Json(out))
}

async fn get_user_by_id(State(state): State<AppState>, Path(id): Path<String>) -> ApiResult<Json<User>> {
    let oid = ObjectId::parse_str(&id).map_err(|_| ApiError::BadRequest("invalid id".into()))?;
    let coll = state.mongodb.collection::<User>("users");
    let found = coll.find_one(doc!{"_id": oid}, None).await.map_err(|e| ApiError::Internal(e.to_string()))?;
    let user = found.ok_or_else(|| ApiError::NotFound("user not found".into()))?;
    Ok(Json(user))
}

async fn update_user(State(state): State<AppState>, Path(id): Path<String>, Json(payload): Json<UpdateUserDto>) -> ApiResult<Json<User>> {
    let oid = ObjectId::parse_str(&id).map_err(|_| ApiError::BadRequest("invalid id".into()))?;
    let coll = state.mongodb.collection::<User>("users");
    let mut update = doc!{};
    if let Some(v) = payload.name { update.insert("name", v); }
    if let Some(v) = payload.lastname { update.insert("lastname", v); }
    if let Some(v) = payload.celular { update.insert("celular", v); }
    let res = coll
        .find_one_and_update(doc!{"_id": oid}, doc!{"$set": update}, None)
        .await
        .map_err(|e| ApiError::Internal(e.to_string()))?;
    let user = res.ok_or_else(|| ApiError::NotFound("user not found".into()))?;
    Ok(Json(user))
}


