#pragma once
#include <mongocxx/client.hpp>
#include <mongocxx/database.hpp>
#include <mongocxx/instance.hpp>
#include <mongocxx/uri.hpp>
#include <sw/redis++/redis++.h>
#include <memory>

class Database {
private:
    static std::unique_ptr<mongocxx::instance> mongo_instance;
    static std::unique_ptr<mongocxx::client> mongo_client;
    static std::unique_ptr<mongocxx::database> mongo_db;
    static std::unique_ptr<sw::redis::Redis> redis_client;

public:
    static void initialize();
    static mongocxx::database& mongodb();
    static sw::redis::Redis& redis();
    static void cleanup();
};
