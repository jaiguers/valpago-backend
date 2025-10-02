#include "security.hpp"
#include <argon2.h>
#include <random>
#include <vector>
#include <iomanip>
#include <sstream>

std::string Security::Password::hash(const std::string& password) {
    const size_t hash_len = 32;
    const size_t salt_len = 16;
    const uint32_t t_cost = 2;      // 2-pass computation
    const uint32_t m_cost = 65536;  // 64 MB memory usage
    const uint32_t parallelism = 1; // number of threads
    
    // Generate random salt
    std::vector<uint8_t> salt(salt_len);
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<> dis(0, 255);
    
    for (auto& byte : salt) {
        byte = static_cast<uint8_t>(dis(gen));
    }
    
    // Hash password
    std::vector<uint8_t> hash_output(hash_len);
    
    int result = argon2i_hash_raw(
        t_cost, m_cost, parallelism,
        password.c_str(), password.length(),
        salt.data(), salt_len,
        hash_output.data(), hash_len
    );
    
    if (result != ARGON2_OK) {
        throw std::runtime_error("Argon2 hashing failed");
    }
    
    // Encode as hex string with salt
    std::stringstream ss;
    for (auto byte : salt) {
        ss << std::hex << std::setw(2) << std::setfill('0') << static_cast<int>(byte);
    }
    ss << ":";
    for (auto byte : hash_output) {
        ss << std::hex << std::setw(2) << std::setfill('0') << static_cast<int>(byte);
    }
    
    return ss.str();
}

bool Security::Password::verify(const std::string& hash, const std::string& password) {
    auto colon_pos = hash.find(':');
    if (colon_pos == std::string::npos) return false;
    
    std::string salt_hex = hash.substr(0, colon_pos);
    std::string hash_hex = hash.substr(colon_pos + 1);
    
    if (salt_hex.length() != 32 || hash_hex.length() != 64) return false;
    
    // Convert hex to bytes
    std::vector<uint8_t> salt(16);
    std::vector<uint8_t> stored_hash(32);
    
    for (size_t i = 0; i < 16; ++i) {
        salt[i] = static_cast<uint8_t>(std::stoi(salt_hex.substr(i * 2, 2), nullptr, 16));
    }
    
    for (size_t i = 0; i < 32; ++i) {
        stored_hash[i] = static_cast<uint8_t>(std::stoi(hash_hex.substr(i * 2, 2), nullptr, 16));
    }
    
    // Hash the provided password with the same salt
    std::vector<uint8_t> computed_hash(32);
    
    int result = argon2i_hash_raw(
        2, 65536, 1,
        password.c_str(), password.length(),
        salt.data(), salt.size(),
        computed_hash.data(), computed_hash.size()
    );
    
    if (result != ARGON2_OK) return false;
    
    // Compare hashes
    return computed_hash == stored_hash;
}
