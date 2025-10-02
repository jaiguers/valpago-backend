#include "database.hpp"
#include "config.hpp"
#include <iostream>
#include <stdexcept>

std::unique_ptr<mongocxx::instance> Database::mongo_instance = nullptr;
std::unique_ptr<mongocxx::client> Database::mongo_client = nullptr;
std::unique_ptr<mongocxx::database> Database::mongo_db = nullptr;

void Database::initialize() {
    try {
        // Initialize MongoDB
        mongo_instance = std::make_unique<mongocxx::instance>();
        mongocxx::uri uri{config.mongodb_uri()};
        mongo_client = std::make_unique<mongocxx::client>(uri);
        mongo_db = std::make_unique<mongocxx::database>((*mongo_client)[config.mongodb_db()]);
        
        std::cout << "MongoDB connected successfully\n";
        
    } catch (const std::exception& e) {
        throw std::runtime_error("Failed to initialize MongoDB: " + std::string(e.what()));
    }
}

mongocxx::database& Database::mongodb() {
    if (!mongo_db) {
        throw std::runtime_error("MongoDB not initialized");
    }
    return *mongo_db;
}

void Database::cleanup() {
    mongo_db.reset();
    mongo_client.reset();
    mongo_instance.reset();
}
