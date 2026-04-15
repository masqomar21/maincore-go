# MainCore Backend (Golang)

A high-performance backend application migrated from TypeScript/Express to Golang using the Gin framework.

## 🚀 Features

- **Fast HTTP Router**: Powered by [Gin Web Framework](https://github.com/gin-gonic/gin).
- **Graceful Shutdown**: Handles OS signals for clean server and background worker termination.
- **Dynamic Port Detection**: Automatically finds and uses an available port if the default is occupied.
- **Database Management**: Integrated GORM with PostgreSQL.
- **Background Tasks**: Redis-powered task queue using [Asynq](https://github.com/hibiken/asynq).
- **Real-time Communication**: WebSockets via Socket.IO.
- **File Storage**: AWS S3 compatible storage integration.
- **Hot Reload**: Seamless development workflow with `air`.

## 🛠 Tech Stack

- **Core**: Go (Golang)
- **Framework**: Gin
- **ORM**: GORM (PostgreSQL)
- **Cache & Queue**: Redis (Asynq)
- **Real-time**: Socket.IO (v2 Protocol)
- **Storage**: AWS S3 SDK v2

## 📂 Project Structure

```text
.
├── cmd/                # Entry points for various commands
│   ├── api/            # Main API Server
│   ├── migrate/        # Database Migration Tool
│   └── seed/           # Database Seeder Tool
├── config/             # Configuration & DB initialization
├── controllers/        # Request handlers
├── middlewares/        # Custom Gin middlewares
├── models/             # GORM models & migration logic
├── routes/             # Route definitions
├── services/           # External services (S3, Socket, Queue)
├── utilities/          # Helper functions (JWT, Responses)
├── main.go             # Root entry (legacy/alternative)
└── .air.toml           # Hot reload configuration
```

## ⚙️ Environment Variables

Create a `.env` file in the root directory:

```env
PORT=3000
DATABASE_URL=postgres://user:password@localhost:5432/dbname
REDIS_HOST=localhost
REDIS_PORT=6379
AUTH_JWT_SECRET=your_jwt_secret
GIN_MODE=debug # or release

# AWS S3
S3_ENDPOINT=
S3_BUCKET=
S3_REGION=
S3_ACCESS_KEY_ID=
S3_SECRET_ACCESS_KEY=
```

## 🚀 Getting Started

### Prerequisites

- Go 1.25+
- PostgreSQL
- Redis
- [Air](https://github.com/cosmtrek/air) (Optional, for hot-reload)

### Running in Development

Use `air` for automatic rebuilding:

```bash
air
```

### Manual Execution

```bash
# Run API Server
go run cmd/api/main.go

# Run Migrations
go run cmd/migrate/main.go

# Run Seeders
go run cmd/seed/main.go
```

## 🏗 Building for Production

Build the binaries for deployment:

```bash
# API Server
go build -o server ./cmd/api

# Migration Tool
go build -o migrate ./cmd/migrate

# Seeder Tool
go build -o seed ./cmd/seed
```

## 📡 Socket.IO Support

This project supports **Socket.IO v4** (and v3). 
When connecting using tools like Postman, Firecamp, or modern web clients, you can use the default settings (**v4 / EIO=4**).

## 📜 License

[MIT](LICENSE)
