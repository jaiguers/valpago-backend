#include <crow.h>
#include <iostream>
#include <csignal>
#include <thread>

#include "config.hpp"
#include "database.hpp"
#include "sse.hpp"
#include "errors.hpp"

// Forward declarations for route setup functions
namespace Routes {
    void setup_user_routes(crow::SimpleApp& app);
    void setup_auth_routes(crow::SimpleApp& app);
    void setup_transaction_routes(crow::SimpleApp& app);
    void setup_sse_routes(crow::SimpleApp& app);
}

// Forward declaration for worker
void start_redis_worker();

// Global flag for graceful shutdown
std::atomic<bool> shutdown_requested{false};

void signal_handler(int signal) {
    std::cout << "\nShutdown signal received (" << signal << "). Gracefully shutting down...\n";
    shutdown_requested.store(true);
}

int main() {
    try {
        std::cout << "Starting ValPago Backend (C++ version)\n";
        
        // Install signal handlers
        std::signal(SIGINT, signal_handler);
        std::signal(SIGTERM, signal_handler);
        
        // Initialize database connections
        std::cout << "Initializing database connections...\n";
        Database::initialize();
        
        // Create Crow app
        crow::SimpleApp app;
        
        // Enable CORS
        auto& cors = app.get_middleware<crow::CORSHandler>();
        cors.global()
            .headers("Content-Type", "Authorization", config.api_key_header_name())
            .methods("GET"_method, "POST"_method, "PUT"_method, "OPTIONS"_method);
        
        if (config.allowed_origins() != "*") {
            cors.origin(config.allowed_origins());
        } else {
            cors.origin("*");
        }
        
        // Setup routes
        std::cout << "Setting up routes...\n";
        
        // Health check
        CROW_ROUTE(app, "/health").methods("GET"_method)([]() {
            return crow::response(200, "ok");
        });
        
        Routes::setup_user_routes(app);
        Routes::setup_auth_routes(app);
        Routes::setup_transaction_routes(app);
        Routes::setup_sse_routes(app);
        
        // Start Redis worker in background thread
        std::cout << "Starting Redis worker...\n";
        std::thread worker_thread(start_redis_worker);
        worker_thread.detach();
        
        // Start server
        int port = config.server_port();
        std::cout << "Server listening on port " << port << "\n";
        std::cout << "Health endpoint: http://localhost:" << port << "/health\n";
        std::cout << "SSE endpoint: http://localhost:" << port << "/api/sse\n";
        
        app.port(port).multithreaded().run();
        
    } catch (const std::exception& e) {
        std::cerr << "Fatal error: " << e.what() << std::endl;
        return 1;
    }
    
    // Cleanup
    std::cout << "Cleaning up...\n";
    SSEManager::instance().shutdown();
    Database::cleanup();
    
    return 0;
}
