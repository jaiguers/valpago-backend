#include "security.hpp"
#include "config.hpp"
#include <chrono>

std::string Security::JWT::create_token(const std::string& subject) {
    auto now = std::chrono::system_clock::now();
    auto exp = now + std::chrono::hours(config.jwt_exp_hours());
    
    auto token = jwt::create()
        .set_issuer("valpago-backend")
        .set_type("JWT")
        .set_subject(subject)
        .set_issued_at(now)
        .set_expires_at(exp)
        .sign(jwt::algorithm::hs256{config.jwt_secret()});
    
    return token;
}

std::optional<std::string> Security::JWT::validate_token(const std::string& token) {
    try {
        auto verifier = jwt::verify()
            .allow_algorithm(jwt::algorithm::hs256{config.jwt_secret()})
            .with_issuer("valpago-backend");
        
        auto decoded = jwt::decode(token);
        verifier.verify(decoded);
        
        return decoded.get_subject();
    } catch (const std::exception&) {
        return std::nullopt;
    }
}
