#include "security.hpp"
#include "database.hpp"
#include <bsoncxx/builder/stream/document.hpp>
#include <bsoncxx/json.hpp>

bool Security::ApiKey::validate(const std::string& key) {
    try {
        auto& db = Database::mongodb();
        auto collection = db["api_keys"];
        
        bsoncxx::builder::stream::document filter_builder;
        filter_builder << "key" << key << "active" << true;
        
        auto result = collection.find_one(filter_builder.view());
        return result.has_value();
        
    } catch (const std::exception&) {
        return false;
    }
}
