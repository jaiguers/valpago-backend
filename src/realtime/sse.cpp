#include "sse.hpp"
#include <algorithm>
#include <sstream>

SSEManager& SSEManager::instance() {
    static SSEManager instance;
    return instance;
}

void SSEManager::add_connection(crow::response::stream_handler&& handler) {
    std::lock_guard<std::mutex> lock(connections_mutex);
    connections.emplace_back(std::make_shared<SSEConnection>(std::move(handler)));
}

void SSEManager::broadcast(const std::string& data, const std::string& event) {
    std::lock_guard<std::mutex> lock(connections_mutex);
    
    std::stringstream sse_message;
    sse_message << "event: " << event << "\n";
    sse_message << "data: " << data << "\n\n";
    
    std::string message = sse_message.str();
    
    for (auto& conn : connections) {
        if (conn->active.load()) {
            try {
                conn->stream.write(message);
                conn->stream.flush();
            } catch (const std::exception&) {
                conn->active.store(false);
            }
        }
    }
}

void SSEManager::cleanup_closed_connections() {
    std::lock_guard<std::mutex> lock(connections_mutex);
    connections.erase(
        std::remove_if(connections.begin(), connections.end(),
            [](const std::shared_ptr<SSEConnection>& conn) {
                return !conn->active.load();
            }),
        connections.end()
    );
}

void SSEManager::shutdown() {
    running.store(false);
    std::lock_guard<std::mutex> lock(connections_mutex);
    
    for (auto& conn : connections) {
        conn->active.store(false);
        try {
            conn->stream.end();
        } catch (const std::exception&) {
            // Ignore errors during shutdown
        }
    }
    connections.clear();
}

crow::response SSEManager::handle_sse_request(const crow::request& req) {
    crow::response res;
    res.set_header("Content-Type", "text/event-stream");
    res.set_header("Cache-Control", "no-cache");
    res.set_header("Connection", "keep-alive");
    res.set_header("Access-Control-Allow-Origin", "*");
    res.set_header("Access-Control-Allow-Headers", "Cache-Control");
    
    return crow::response::stream([this](crow::response::stream_handler& handler) {
        // Send initial connection message
        handler.write("event: connected\n");
        handler.write("data: {\"message\": \"SSE connection established\"}\n\n");
        handler.flush();
        
        // Add to active connections
        add_connection(std::move(handler));
        
        // Keep connection alive (handled by connection cleanup)
    });
}

namespace Routes {
    void setup_sse_routes(crow::SimpleApp& app) {
        CROW_ROUTE(app, "/api/sse").methods("GET"_method)([](const crow::request& req) {
            return SSEManager::instance().handle_sse_request(req);
        });
    }
}
