# ValPago Backend - C++ Version

API REST en C++ para la aplicación ValPago, migrada desde Rust con arquitectura modular y notificaciones en tiempo real via Server-Sent Events.

## Características

- **Framework Web**: Crow (C++ micro-framework)
- **Base de Datos**: MongoDB con driver mongocxx
- **Cache/Streams**: Redis con redis-plus-plus
- **Autenticación**: JWT con jwt-cpp
- **Seguridad**: Hash de passwords con Argon2
- **Tiempo Real**: Server-Sent Events (SSE) para notificaciones
- **Worker**: Procesamiento en background de Redis Streams
- **Arquitectura**: Clean code con separación de responsabilidades

## Dependencias

### Requeridas
- CMake 3.20+
- C++20 compatible compiler (GCC 10+, Clang 12+, MSVC 2019+)
- MongoDB C++ Driver (mongocxx)
- Redis C++ client (redis-plus-plus)
- Crow framework
- jwt-cpp
- nlohmann/json
- Argon2 library
- OpenSSL

### Instalación de dependencias (Ubuntu/Debian)
```bash
# Dependencias del sistema
sudo apt update
sudo apt install cmake build-essential pkg-config libssl-dev

# MongoDB C++ driver
sudo apt install libmongocxx-dev libbsoncxx-dev

# Redis client
sudo apt install libhiredis-dev

# Argon2
sudo apt install libargon2-dev

# Crow, jwt-cpp, nlohmann/json - via vcpkg o compilación manual
```

### Usando vcpkg (recomendado)
```bash
git clone https://github.com/Microsoft/vcpkg.git
./vcpkg/bootstrap-vcpkg.sh
./vcpkg/vcpkg install crow jwt-cpp nlohmann-json redis-plus-plus
```

## Compilación

```bash
mkdir build && cd build
cmake .. -DCMAKE_TOOLCHAIN_FILE=[path-to-vcpkg]/scripts/buildsystems/vcpkg.cmake
make -j$(nproc)
```

## Configuración

Crea un archivo `.env` en la raíz del proyecto:

```env
MONGODB_URI=mongodb://localhost:27017
MONGODB_DB=valpago

REDIS_URL=redis://127.0.0.1:6379
REDIS_STREAM_NAMESPACE=valpago:transactions
REDIS_CONSUMER_GROUP=valpago:cg
REDIS_CONSUMER_NAME=worker-1

JWT_SECRET=super_secret_change_me
JWT_EXP_HOURS=24

API_KEY_HEADER_NAME=x-api-key
SERVER_PORT=8080
ALLOWED_ORIGINS=*
```

Para Redis Cloud con TLS:
```env
REDIS_URL=rediss://:password@host:port
```

## Ejecución

```bash
./valpago-backend
```

El servidor iniciará en `http://localhost:8080`

## Endpoints

### Health Check
- `GET /health` - Estado del servidor

### Usuarios (`/api/users`)
- `POST /api/users` - Crear usuario
- `GET /api/users` - Listar usuarios
- `GET /api/users/{id}` - Obtener usuario por ID
- `PUT /api/users/{id}` - Actualizar usuario

### Autenticación (`/api/auth`)
- `POST /api/auth/login` - Login con email/password

### Transacciones (`/api/transactions`)
- `POST /api/transactions/create` - Crear transacción (requiere API Key)
- `GET /api/transactions` - Listar transacciones
- `PUT /api/transactions/{id}/status` - Actualizar estado

### Notificaciones en Tiempo Real
- `GET /api/sse` - Server-Sent Events stream

## Seguridad

### Headers requeridos:
- JWT: `Authorization: Bearer <token>`
- API Key: `x-api-key: <key>` (configurable)

### Colecciones MongoDB:
- `users`: Usuarios con passwords hasheados (Argon2)
- `transactions`: Transacciones con estados
- `api_keys`: Claves API válidas `{key: string, active: bool}`

## Arquitectura

```
src/
├── main.cpp                 # Punto de entrada
├── config/
│   └── config.cpp          # Carga de configuración .env
├── database/
│   ├── mongodb.cpp         # Conexión MongoDB
│   └── redis.cpp           # Conexión Redis
├── security/
│   ├── password.cpp        # Hash Argon2
│   ├── jwt.cpp             # Manejo JWT
│   └── apikey.cpp          # Validación API Keys
├── routes/
│   ├── users.cpp           # Endpoints usuarios
│   ├── auth.cpp            # Endpoints autenticación
│   └── transactions.cpp    # Endpoints transacciones
├── realtime/
│   └── sse.cpp             # Server-Sent Events
├── worker/
│   └── redis_worker.cpp    # Worker Redis Streams
└── errors/
    └── api_error.cpp       # Manejo de errores
```

## Redis Streams

- **Stream**: `valpago:transactions`
- **Consumer Group**: `valpago:cg`
- **Worker**: Procesa mensajes y envía notificaciones SSE
- **Estados**: `PENDING` → `REVIEW` → `APPROVED`/`REJECTED`

## Desarrollo

### Compilación debug:
```bash
cmake .. -DCMAKE_BUILD_TYPE=Debug
make
```

### Testing:
```bash
# Health check
curl http://localhost:8080/health

# SSE connection
curl -N http://localhost:8080/api/sse

# Create user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","lastname":"User","email":"test@test.com","password":"123456","celular":"123456789"}'
```

## Migración desde Rust

Esta versión mantiene la misma API y funcionalidad que la versión original en Rust, con las siguientes diferencias:

- **WebSockets → Server-Sent Events**: Más simple para notificaciones unidireccionales
- **Axum → Crow**: Framework C++ ligero y eficiente
- **Wither ODM → mongocxx nativo**: Driver oficial MongoDB
- **tokio → std::thread**: Threading nativo C++

## Performance

- **Memoria**: Menor overhead vs Rust (sin garbage collector)
- **Latencia**: Comparable o mejor que Rust para I/O intensivo
- **Throughput**: Excelente con multithreading nativo
- **SSE**: Menor overhead que WebSockets para notificaciones push