use chrono::{Duration, Utc};
use jsonwebtoken::{encode, decode, Header, Algorithm, Validation, EncodingKey, DecodingKey};
use serde::{Serialize, Deserialize};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Claims {
    pub sub: String,
    pub exp: i64,
}

pub fn create_jwt(subject: &str) -> anyhow::Result<String> {
    let secret = std::env::var("JWT_SECRET").expect("JWT_SECRET is required");
    let hours: i64 = std::env::var("JWT_EXP_HOURS").ok().and_then(|v| v.parse().ok()).unwrap_or(24);
    let exp = Utc::now() + Duration::hours(hours);
    let claims = Claims { sub: subject.to_string(), exp: exp.timestamp() };
    let token = encode(&Header::new(Algorithm::HS256), &claims, &EncodingKey::from_secret(secret.as_bytes()))?;
    Ok(token)
}

pub fn validate_jwt(token: &str) -> anyhow::Result<Claims> {
    let secret = std::env::var("JWT_SECRET").expect("JWT_SECRET is required");
    let data = decode::<Claims>(token, &DecodingKey::from_secret(secret.as_bytes()), &Validation::new(Algorithm::HS256))?;
    Ok(data.claims)
}

