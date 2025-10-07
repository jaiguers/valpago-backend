use axum::{extract::{State}, routing::{get, post, put}, Json, Router};
use axum::http::HeaderMap;
use futures::TryStreamExt;
use mongodb::bson::{doc, oid::ObjectId};
use serde::{Deserialize, Serialize};

use crate::{errors::{ApiError, ApiResult}, routes::AppState, security::{apikey::validate_api_key}};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub enum TransactionStatus {
    PENDING="pending",
    REVIEW="review",
    APPROVED="approved",
    REJECTED="rejected",
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Transaction {
    #[serde(rename = "_id", skip_serializing_if = "Option::is_none")]
    pub id: Option<ObjectId>,
    pub metodo_pago: String,
    pub monto: String,
    pub cuenta_consignacion: String,
    pub referencia: String,
    pub cuenta_origen: String,
    pub beneficiario: String,
    pub tel_whatsapp_send: String,
    pub estado: TransactionStatus,
    pub url_soporte: String,
    pub date: String,
}

#[derive(Debug, Deserialize)]
pub struct CreateTransactionDto {
    pub metodo_pago: String,
    pub monto: String,
    pub cuenta_consignacion: String,
    pub referencia: String,
    pub cuenta_origen: String,
    pub beneficiario: String,
    pub tel_whatsapp_send: String,
    pub estado: String,
    pub url_soporte: String,
    pub date: String,
}

#[derive(Debug, Deserialize)]
pub struct UpdateStatusDto {
    pub estado: String,
}

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/create", post(create_transaction))
        .route("/", get(list_transactions))
        .route("/:id/status", put(update_status))
}

async fn create_transaction(State(state): State<AppState>, headers: HeaderMap, Json(payload): Json<CreateTransactionDto>) -> ApiResult<Json<Transaction>> {
    let header_name = std::env::var("API_KEY_HEADER_NAME").unwrap_or_else(|_| "x-api-key".into());
    let api_key = headers.get(&header_name).and_then(|v| v.to_str().ok()).ok_or_else(|| ApiError::Unauthorized("missing API key".into()))?;
    if !validate_api_key(&state.mongodb, api_key).await.map_err(|e| ApiError::Internal(e.to_string()))? {
        return Err(ApiError::Forbidden("invalid API key".into()));
    }

    let status = TransactionStatus::PENDING;
    let mut tx = Transaction {
        id: None,
        metodo_pago: payload.metodo_pago,
        monto: payload.monto,
        cuenta_consignacion: payload.cuenta_consignacion,
        referencia: payload.referencia,
        cuenta_origen: payload.cuenta_origen,
        beneficiario: payload.beneficiario,
        tel_whatsapp_send: payload.tel_whatsapp_send,
        estado: status,
        url_soporte: payload.url_soporte,
        date: payload.date,
    };

    // Save DB
    let coll = state.mongodb.collection::<Transaction>("transactions");
    coll.insert_one(&tx, None).await.map_err(|e| ApiError::Internal(e.to_string()))?;

    // Publish to Redis Stream
    publish_transaction(&state, &tx).await.map_err(|e| ApiError::Internal(e.to_string()))?;

    Ok(Json(tx))
}

async fn list_transactions(State(state): State<AppState>) -> ApiResult<Json<Vec<Transaction>>> {
    let coll = state.mongodb.collection::<Transaction>("transactions");
    let mut cursor = coll.find(None, None).await.map_err(|e| ApiError::Internal(e.to_string()))?;
    let mut out = Vec::new();
    while let Some(doc) = cursor.try_next().await.map_err(|e| ApiError::Internal(e.to_string()))? { out.push(doc); }
    Ok(Json(out))
}

async fn update_status(State(state): State<AppState>, axum::extract::Path(id): axum::extract::Path<String>, Json(payload): Json<UpdateStatusDto>) -> ApiResult<Json<Transaction>> {
    let oid = ObjectId::parse_str(&id).map_err(|_| ApiError::BadRequest("invalid id".into()))?;
    let coll = state.mongodb.collection::<Transaction>("transactions");
    let res = coll
        .find_one_and_update(doc!{"_id": oid}, doc!{"$set": {"estado": &payload.estado}}, None)
        .await
        .map_err(|e| ApiError::Internal(e.to_string()))?;
    let tx = res.ok_or_else(|| ApiError::NotFound("transaction not found".into()))?;

    // Notify realtime
    state.notifier.broadcast(&serde_json::to_string(&tx).unwrap());
    Ok(Json(tx))
}

async fn publish_transaction(state: &AppState, tx: &Transaction) -> anyhow::Result<()> {
    use redis::{AsyncCommands, streams::StreamMaxlen};
    let mut conn = state.redis.get_async_connection().await?;
    let ns = crate::db::redis::stream_namespace();
    let payload = serde_json::to_string(tx)?;
    let _: String = redis::cmd("XADD")
        .arg(&ns)
        .arg("MAXLEN")
        .arg("~")
        .arg(10000)
        .arg("*")
        .arg("payload")
        .arg(payload)
        .query_async(&mut conn)
        .await?;
    Ok(())
}

