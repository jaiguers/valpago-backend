#pragma once
#include <string>
#include <nlohmann/json.hpp>
#include <bsoncxx/document/value.hpp>
#include <bsoncxx/oid.hpp>

enum class TransactionStatus {
    PENDING = "pending",
    REVIEW = "review",
    APPROVED = "approved",
    REJECTED = "rejected"
};

struct User {
    std::string id;
    std::string name;
    std::string lastname;
    std::string email;
    std::string password;
    std::string celular;
    
    nlohmann::json to_json(bool include_password = false) const;
    static User from_json(const nlohmann::json& j);
    static User from_bson(const bsoncxx::document::view& doc);
    bsoncxx::document::value to_bson() const;
};

struct Transaction {
    std::string id;
    std::string metodo_pago;
    std::string monto;
    std::string cuenta_consignacion;
    std::string referencia;
    std::string cuenta_origen;
    std::string beneficiario;
    std::string tel_whatsapp_send;
    TransactionStatus estado;
    std::string url_soporte;
    std::string date;
    
    nlohmann::json to_json() const;
    static Transaction from_json(const nlohmann::json& j);
    static Transaction from_bson(const bsoncxx::document::view& doc);
    bsoncxx::document::value to_bson() const;
};

struct CreateUserDto {
    std::string name;
    std::string lastname;
    std::string email;
    std::string password;
    std::string celular;
    
    static CreateUserDto from_json(const nlohmann::json& j);
};

struct UpdateUserDto {
    std::string name;
    std::string lastname;
    std::string celular;
    
    static UpdateUserDto from_json(const nlohmann::json& j);
    bool has_name() const { return !name.empty(); }
    bool has_lastname() const { return !lastname.empty(); }
    bool has_celular() const { return !celular.empty(); }
};

struct LoginDto {
    std::string email;
    std::string password;
    
    static LoginDto from_json(const nlohmann::json& j);
};

struct CreateTransactionDto {
    std::string metodo_pago;
    std::string monto;
    std::string cuenta_consignacion;
    std::string referencia;
    std::string cuenta_origen;
    std::string beneficiario;
    std::string tel_whatsapp_send;
    std::string url_soporte;
    std::string date;
    
    static CreateTransactionDto from_json(const nlohmann::json& j);
};

struct UpdateStatusDto {
    std::string estado;
    
    static UpdateStatusDto from_json(const nlohmann::json& j);
};

// Helper functions
std::string transaction_status_to_string(TransactionStatus status);
TransactionStatus transaction_status_from_string(const std::string& str);
