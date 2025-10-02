#include "errors.hpp"

ApiError::ApiError(ApiErrorType type, const std::string& message) 
    : type(type), message(message) {}

crow::response ApiError::to_response() const {
    auto json_body = to_json();
    crow::response res(status_code(), json_body.dump());
    res.set_header("Content-Type", "application/json");
    return res;
}

nlohmann::json ApiError::to_json() const {
    return nlohmann::json{
        {"error", error_name()},
        {"message", message}
    };
}

int ApiError::status_code() const {
    switch (type) {
        case ApiErrorType::BAD_REQUEST: return 400;
        case ApiErrorType::UNAUTHORIZED: return 401;
        case ApiErrorType::FORBIDDEN: return 403;
        case ApiErrorType::NOT_FOUND: return 404;
        case ApiErrorType::CONFLICT: return 409;
        case ApiErrorType::INTERNAL_ERROR: return 500;
        default: return 500;
    }
}

std::string ApiError::error_name() const {
    switch (type) {
        case ApiErrorType::BAD_REQUEST: return "bad_request";
        case ApiErrorType::UNAUTHORIZED: return "unauthorized";
        case ApiErrorType::FORBIDDEN: return "forbidden";
        case ApiErrorType::NOT_FOUND: return "not_found";
        case ApiErrorType::CONFLICT: return "conflict";
        case ApiErrorType::INTERNAL_ERROR: return "internal_error";
        default: return "unknown_error";
    }
}

ApiError ApiError::bad_request(const std::string& message) {
    return ApiError(ApiErrorType::BAD_REQUEST, message);
}

ApiError ApiError::unauthorized(const std::string& message) {
    return ApiError(ApiErrorType::UNAUTHORIZED, message);
}

ApiError ApiError::forbidden(const std::string& message) {
    return ApiError(ApiErrorType::FORBIDDEN, message);
}

ApiError ApiError::not_found(const std::string& message) {
    return ApiError(ApiErrorType::NOT_FOUND, message);
}

ApiError ApiError::conflict(const std::string& message) {
    return ApiError(ApiErrorType::CONFLICT, message);
}

ApiError ApiError::internal_error(const std::string& message) {
    return ApiError(ApiErrorType::INTERNAL_ERROR, message);
}
