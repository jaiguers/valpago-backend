use std::task::{Context, Poll};
use std::future::Future;
use std::pin::Pin;
use axum::{body::Body, http::{Request, StatusCode}, response::Response};
use tower::{Layer, Service};

use crate::{errors::ApiError, routes::AppState, security::jwt};

#[derive(Clone)]
pub struct JwtLayer;

impl<S> Layer<S> for JwtLayer {
    type Service = JwtMiddleware<S>;
    fn layer(&self, inner: S) -> Self::Service { JwtMiddleware { inner } }
}

#[derive(Clone)]
pub struct JwtMiddleware<S> { inner: S }

impl<S> Service<Request<Body>> for JwtMiddleware<S>
where
    S: Service<Request<Body>, Response = Response> + Clone + Send + 'static,
    S::Future: Send + 'static,
{
    type Response = S::Response;
    type Error = S::Error;
    type Future = S::Future;

    fn poll_ready(&mut self, cx: &mut Context<'_>) -> Poll<Result<(), Self::Error>> {
        self.inner.poll_ready(cx)
    }

    fn call(&mut self, req: Request<Body>) -> Self::Future {
        // Allow public routes
        let path = req.uri().path().to_string();
        if path.starts_with("/api/auth/") || path == "/health" || (path.starts_with("/api/transactions/create")) {
            return self.inner.call(req);
        }
        if let Some(token) = req.headers().get("authorization").and_then(|v| v.to_str().ok()) {
            let token = token.strip_prefix("Bearer ").unwrap_or(token);
            if jwt::validate_jwt(token).is_ok() {
                return self.inner.call(req);
            }
        }
        let mut res = Response::new(Body::from(serde_json::to_string(&crate::errors::ApiErrorBody{error:"unauthorized".into(), message:"missing or invalid token".into()}).unwrap()));
        *res.status_mut() = StatusCode::UNAUTHORIZED;
        res
    }
}

#[derive(Clone)]
pub struct ApiKeyLayer;

impl<S> Layer<S> for ApiKeyLayer {
    type Service = ApiKeyMiddleware<S>;
    fn layer(&self, inner: S) -> Self::Service { ApiKeyMiddleware { inner } }
}

#[derive(Clone)]
pub struct ApiKeyMiddleware<S> { inner: S }

impl<S> Service<Request<Body>> for ApiKeyMiddleware<S>
where
    S: Service<Request<Body>, Response = Response> + Clone + Send + 'static,
    S::Future: Send + 'static,
{
    type Response = S::Response;
    type Error = S::Error;
    type Future = S::Future;

    fn poll_ready(&mut self, cx: &mut Context<'_>) -> Poll<Result<(), Self::Error>> {
        self.inner.poll_ready(cx)
    }

    fn call(&mut self, req: Request<Body>) -> Self::Future {
        // Only protect transaction creation endpoint by API Key
        let path = req.uri().path().to_string();
        if path == "/api/transactions/create" {
            // Just forward; handler validates API Key in DB for better error message
            return self.inner.call(req);
        }
        self.inner.call(req)
    }
}


