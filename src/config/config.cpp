#include "config.hpp"
#include <fstream>
#include <iostream>
#include <sstream>
#include <stdexcept>

Config config;

Config::Config() {
    load_env_file();
}

void Config::load_env_file() {
    std::ifstream file(".env");
    if (!file.is_open()) {
        std::cerr << "Warning: .env file not found\n";
        return;
    }
    
    std::string line;
    while (std::getline(file, line)) {
        if (line.empty() || line[0] == '#') continue;
        
        auto pos = line.find('=');
        if (pos != std::string::npos) {
            std::string key = line.substr(0, pos);
            std::string value = line.substr(pos + 1);
            env_vars[key] = value;
        }
    }
}

std::string Config::get(const std::string& key, const std::string& default_value) const {
    auto it = env_vars.find(key);
    return (it != env_vars.end()) ? it->second : default_value;
}

int Config::get_int(const std::string& key, int default_value) const {
    auto value = get(key);
    if (value.empty()) return default_value;
    try {
        return std::stoi(value);
    } catch (const std::exception&) {
        return default_value;
    }
}

bool Config::get_bool(const std::string& key, bool default_value) const {
    auto value = get(key);
    if (value.empty()) return default_value;
    return value == "true" || value == "1" || value == "yes";
}

std::string Config::mongodb_uri() const {
    auto uri = get("MONGODB_URI");
    if (uri.empty()) {
        throw std::runtime_error("MONGODB_URI is required");
    }
    return uri;
}

std::string Config::mongodb_db() const {
    return get("MONGODB_DB", "valpago");
}

std::string Config::redis_url() const {
    auto url = get("REDIS_URL");
    if (url.empty()) {
        throw std::runtime_error("REDIS_URL is required");
    }
    return url;
}

std::string Config::redis_stream_namespace() const {
    return get("REDIS_STREAM_NAMESPACE", "valpago:transactions");
}

std::string Config::redis_consumer_group() const {
    return get("REDIS_CONSUMER_GROUP", "valpago:cg");
}

std::string Config::redis_consumer_name() const {
    return get("REDIS_CONSUMER_NAME", "worker-1");
}

std::string Config::jwt_secret() const {
    auto secret = get("JWT_SECRET");
    if (secret.empty()) {
        throw std::runtime_error("JWT_SECRET is required");
    }
    return secret;
}

int Config::jwt_exp_hours() const {
    return get_int("JWT_EXP_HOURS", 24);
}

std::string Config::api_key_header_name() const {
    return get("API_KEY_HEADER_NAME", "x-api-key");
}

int Config::server_port() const {
    return get_int("SERVER_PORT", 8080);
}

std::string Config::allowed_origins() const {
    return get("ALLOWED_ORIGINS", "*");
}
