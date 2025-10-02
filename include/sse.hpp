#pragma once
#include <crow.h>
#include <string>
#include <memory>
#include <vector>
#include <mutex>
#include <thread>
#include <atomic>

class SSEManager {
private:
    struct SSEConnection {
        crow::response::stream_handler stream;
        std::atomic<bool> active{true};
        
        SSEConnection(crow::response::stream_handler&& handler) 
            : stream(std::move(handler)) {}
    };
    
    std::vector<std::shared_ptr<SSEConnection>> connections;
    std::mutex connections_mutex;
    std::atomic<bool> running{true};

public:
    static SSEManager& instance();
    
    void add_connection(crow::response::stream_handler&& handler);
    void broadcast(const std::string& data, const std::string& event = "message");
    void cleanup_closed_connections();
    void shutdown();
    
    // SSE endpoint handler
    crow::response handle_sse_request(const crow::request& req);
};

namespace Routes {
    void setup_sse_routes(crow::SimpleApp& app);
}
