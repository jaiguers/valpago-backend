#pragma once
#include <string>
#include <nlohmann/json.hpp>
#include <crow.h>

enum class ApiErrorType {
    BAD_REQUEST,
    UNAUTHORIZED,
    FORBIDDEN,
    NOT_FOUND,
    CONFLICT,
    INTERNAL_ERROR
};

class ApiError {
private:
    ApiErrorType type;
    std::string message;

public:
    ApiError(ApiErrorType type, const std::string& message);
    
    crow::response to_response() const;
    nlohmann::json to_json() const;
    int status_code() const;
    std::string error_name() const;
    
    static ApiError bad_request(const std::string& message);
    static ApiError unauthorized(const std::string& message);
    static ApiError forbidden(const std::string& message);
    static ApiError not_found(const std::string& message);
    static ApiError conflict(const std::string& message);
    static ApiError internal_error(const std::string& message);
};
