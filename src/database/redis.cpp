#include "database.hpp"
#include "config.hpp"
#include <iostream>
#include <stdexcept>

std::unique_ptr<sw::redis::Redis> Database::redis_client = nullptr;

void Database::initialize() {
    try {
        // Initialize Redis
        sw::redis::ConnectionOptions connection_options;
        connection_options.host = "localhost";  // Parse from redis_url in real implementation
        connection_options.port = 6379;
        
        // Parse Redis URL (simplified - in production parse properly)
        std::string redis_url = config.redis_url();
        if (redis_url.find("rediss://") == 0) {
            connection_options.tls.enabled = true;
        }
        
        redis_client = std::make_unique<sw::redis::Redis>(connection_options);
        
        // Test connection
        redis_client->ping();
        std::cout << "Redis connected successfully\n";
        
    } catch (const std::exception& e) {
        throw std::runtime_error("Failed to initialize Redis: " + std::string(e.what()));
    }
}

sw::redis::Redis& Database::redis() {
    if (!redis_client) {
        throw std::runtime_error("Redis not initialized");
    }
    return *redis_client;
}
