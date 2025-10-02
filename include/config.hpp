#pragma once
#include <string>
#include <unordered_map>

class Config {
private:
    std::unordered_map<std::string, std::string> env_vars;
    void load_env_file();

public:
    Config();
    
    std::string get(const std::string& key, const std::string& default_value = "") const;
    int get_int(const std::string& key, int default_value = 0) const;
    bool get_bool(const std::string& key, bool default_value = false) const;
    
    // Specific getters for common config values
    std::string mongodb_uri() const;
    std::string mongodb_db() const;
    std::string redis_url() const;
    std::string redis_stream_namespace() const;
    std::string redis_consumer_group() const;
    std::string redis_consumer_name() const;
    std::string jwt_secret() const;
    int jwt_exp_hours() const;
    std::string api_key_header_name() const;
    int server_port() const;
    std::string allowed_origins() const;
};

extern Config config;
