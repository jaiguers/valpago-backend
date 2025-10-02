#pragma once
#include <string>
#include <optional>
#include <jwt-cpp/jwt.h>

namespace Security {
    // Password hashing with Argon2
    class Password {
    public:
        static std::string hash(const std::string& password);
        static bool verify(const std::string& hash, const std::string& password);
    };
    
    // JWT token handling
    class JWT {
    public:
        static std::string create_token(const std::string& subject);
        static std::optional<std::string> validate_token(const std::string& token);
    };
    
    // API Key validation
    class ApiKey {
    public:
        static bool validate(const std::string& key);
    };
}
