#include "models.hpp"
#include "database.hpp"
#include "security.hpp"
#include "errors.hpp"
#include <crow.h>
#include <bsoncxx/builder/stream/document.hpp>
#include <bsoncxx/json.hpp>
#include <mongocxx/exception/exception.hpp>

namespace Routes {
    
crow::response create_user(const crow::request& req) {
    try {
        auto json_body = nlohmann::json::parse(req.body);
        auto dto = CreateUserDto::from_json(json_body);
        
        // Hash password
        std::string hashed_password = Security::Password::hash(dto.password);
        
        // Create user
        User user;
        user.name = dto.name;
        user.lastname = dto.lastname;
        user.email = dto.email;
        user.password = hashed_password;
        user.celular = dto.celular;
        
        // Save to database
        auto& db = Database::mongodb();
        auto collection = db["users"];
        
        auto result = collection.insert_one(user.to_bson().view());
        if (!result) {
            return ApiError::internal_error("Failed to create user").to_response();
        }
        
        user.id = result->inserted_id().get_oid().value.to_string();
        
        return crow::response(201, user.to_json(false).dump());
        
    } catch (const nlohmann::json::exception&) {
        return ApiError::bad_request("Invalid JSON").to_response();
    } catch (const std::exception& e) {
        return ApiError::internal_error(e.what()).to_response();
    }
}

crow::response list_users(const crow::request& req) {
    try {
        auto& db = Database::mongodb();
        auto collection = db["users"];
        
        auto cursor = collection.find({});
        nlohmann::json users = nlohmann::json::array();
        
        for (auto&& doc : cursor) {
            auto user = User::from_bson(doc);
            users.push_back(user.to_json(false));
        }
        
        return crow::response(200, users.dump());
        
    } catch (const std::exception& e) {
        return ApiError::internal_error(e.what()).to_response();
    }
}

crow::response get_user_by_id(const crow::request& req, const std::string& id) {
    try {
        bsoncxx::oid oid{id};
        
        auto& db = Database::mongodb();
        auto collection = db["users"];
        
        bsoncxx::builder::stream::document filter_builder;
        filter_builder << "_id" << oid;
        
        auto result = collection.find_one(filter_builder.view());
        if (!result) {
            return ApiError::not_found("User not found").to_response();
        }
        
        auto user = User::from_bson(result->view());
        return crow::response(200, user.to_json(false).dump());
        
    } catch (const bsoncxx::exception&) {
        return ApiError::bad_request("Invalid user ID").to_response();
    } catch (const std::exception& e) {
        return ApiError::internal_error(e.what()).to_response();
    }
}

crow::response update_user(const crow::request& req, const std::string& id) {
    try {
        bsoncxx::oid oid{id};
        auto json_body = nlohmann::json::parse(req.body);
        auto dto = UpdateUserDto::from_json(json_body);
        
        auto& db = Database::mongodb();
        auto collection = db["users"];
        
        bsoncxx::builder::stream::document filter_builder;
        filter_builder << "_id" << oid;
        
        bsoncxx::builder::stream::document update_builder;
        update_builder << "$set" << bsoncxx::builder::stream::open_document;
        
        if (dto.has_name()) update_builder << "name" << dto.name;
        if (dto.has_lastname()) update_builder << "lastname" << dto.lastname;
        if (dto.has_celular()) update_builder << "celular" << dto.celular;
        
        update_builder << bsoncxx::builder::stream::close_document;
        
        auto result = collection.find_one_and_update(
            filter_builder.view(),
            update_builder.view()
        );
        
        if (!result) {
            return ApiError::not_found("User not found").to_response();
        }
        
        auto user = User::from_bson(result->view());
        return crow::response(200, user.to_json(false).dump());
        
    } catch (const bsoncxx::exception&) {
        return ApiError::bad_request("Invalid user ID").to_response();
    } catch (const nlohmann::json::exception&) {
        return ApiError::bad_request("Invalid JSON").to_response();
    } catch (const std::exception& e) {
        return ApiError::internal_error(e.what()).to_response();
    }
}

void setup_user_routes(crow::SimpleApp& app) {
    CROW_ROUTE(app, "/api/users").methods("POST"_method)(create_user);
    CROW_ROUTE(app, "/api/users").methods("GET"_method)(list_users);
    CROW_ROUTE(app, "/api/users/<string>").methods("GET"_method)(get_user_by_id);
    CROW_ROUTE(app, "/api/users/<string>").methods("PUT"_method)(update_user);
}

} // namespace Routes
