cmake_minimum_required(VERSION 3.20)
project(valpago-backend VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Find packages
find_package(PkgConfig REQUIRED)
find_package(Threads REQUIRED)
find_package(OpenSSL REQUIRED)

# Crow framework
find_package(Crow REQUIRED)

# MongoDB C++ driver
find_package(mongocxx REQUIRED)
find_package(bsoncxx REQUIRED)

# Redis client
find_package(hiredis REQUIRED)
find_package(redis++ REQUIRED)

# JWT library
find_package(jwt-cpp REQUIRED)

# JSON library
find_package(nlohmann_json REQUIRED)

# Argon2 for password hashing
find_package(PkgConfig REQUIRED)
pkg_check_modules(ARGON2 REQUIRED libargon2)

# Include directories
include_directories(${CMAKE_SOURCE_DIR}/include)

# Source files
set(SOURCES
    src/main.cpp
    src/config/config.cpp
    src/database/mongodb.cpp
    src/database/redis.cpp
    src/security/password.cpp
    src/security/jwt.cpp
    src/security/apikey.cpp
    src/routes/users.cpp
    src/routes/auth.cpp
    src/routes/transactions.cpp
    src/realtime/sse.cpp
    src/worker/redis_worker.cpp
    src/errors/api_error.cpp
)

# Executable
add_executable(valpago-backend ${SOURCES})

# Link libraries
target_link_libraries(valpago-backend
    Crow::Crow
    mongo::mongocxx_shared
    mongo::bsoncxx_shared
    hiredis::hiredis
    redis++::redis++
    jwt-cpp::jwt-cpp
    nlohmann_json::nlohmann_json
    OpenSSL::SSL
    OpenSSL::Crypto
    ${ARGON2_LIBRARIES}
    Threads::Threads
)

# Compiler flags
target_compile_options(valpago-backend PRIVATE
    ${ARGON2_CFLAGS_OTHER}
    -Wall -Wextra -O2
)

# Include directories for argon2
target_include_directories(valpago-backend PRIVATE ${ARGON2_INCLUDE_DIRS})
