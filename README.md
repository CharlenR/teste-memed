# Segmentation API

**API de segmentações com worker de processamento de arquivos**

Production-ready Go application for processing CSV files with segmentation data, featuring multi-threaded processing, RESTful API endpoints, and comprehensive database persistence.

## 📋 Table of Contents

- [Architecture](#architecture)
- [Design Principles](#design-principles)
- [Tech Stack](#tech-stack)
- [Local Setup](#local-setup)

---

## 🏗 Architecture

### System Components

The application consists of three main components:

```
┌─────────────────────────────────────────────┐
│         Segmentation API System              │
├─────────────────────────────────────────────┤
│                                             │
│  ┌──────────────────────────────────────┐   │
│  │   API Server (Gin Framework)         │   │
│  │   - Health checks                    │   │
│  │   - User segmentation queries        │   │
│  │   - RESTful endpoints                │   │
│  └──────────────────────────────────────┘   │
│                    ↑ ↓                       │
│  ┌──────────────────────────────────────┐   │
│  │   Service Layer (Business Logic)     │   │
│  │   - Data validation                  │   │
│  │   - Normalization                    │   │
│  │   - UPSERT logic                     │   │
│  └──────────────────────────────────────┘   │
│                    ↑ ↓                       │
│  ┌──────────────────────────────────────┐   │
│  │   Repository Layer (Data Access)     │   │
│  │   - Database operations              │   │
│  │   - GORM ORM integration             │   │
│  └──────────────────────────────────────┘   │
│                    ↑ ↓                       │
│  ┌──────────────────────────────────────┐   │
│  │   CSV Processor (Workers)            │   │
│  │   - Multi-threaded processing        │   │
│  │   - CPU-aware workers                │   │
│  │   - Progress tracking                │   │
│  │   - 3,600+ records/second            │   │
│  └──────────────────────────────────────┘   │
│                    ↑ ↓                       │
│  ┌──────────────────────────────────────┐   │
│  │   MariaDB Database                   │   │
│  │   - Data persistence                 │   │
│  │   - Transaction support              │   │
│  └──────────────────────────────────────┘   │
│                                             │
└─────────────────────────────────────────────┘
```

### Layered Architecture

| Layer | Purpose | Components |
|-------|---------|------------|
| **API Handler** | HTTP request/response | `internal/api/handler/` |
| **Service** | Business logic & validation | `internal/service/` |
| **Repository** | Data abstraction layer | `internal/repository/` |
| **Models** | Data structures | `internal/models/` |
| **Processor** | CSV processing workers | `internal/processor/` |
| **Logger** | Logging utilities | `internal/logger/` |

### Data Flow

**CSV Processing Flow:**
```
CSV File
  ↓
Parser (line-by-line)
  ↓
Worker Pool (N CPU-aware threads)
  ↓
Validation & Normalization
  ↓
Service Layer (Business Logic)
  ↓
UPSERT to Database
  ↓
Logs (dual output: stdout + file)
```

**API Query Flow:**
```
HTTP Request
  ↓
Router (Gin)
  ↓
Handler
  ↓
Service (Business Logic)
  ↓
Repository (GORM)
  ↓
Database Query
  ↓
Response (JSON)
```

---

## 🎯 Design Principles

### 1. **Separation of Concerns**
- Each layer has a single responsibility
- Handler → Service → Repository → Database
- Easy to test and maintain

### 2. **Dependency Injection**
- Services receive dependencies via constructor
- Repository interface allows swapping implementations
- Testable without external dependencies

### 3. **Performance First**
- Streaming CSV processing (no full load into memory)
- CPU-aware worker count (uses `runtime.NumCPU()`)
- Connection pooling with GORM
- Batch operations where applicable

### 4. **Data Integrity**
- UPSERT logic prevents duplicates
- Proper transaction handling
- Input validation before persistence
- Normalization ensures consistency

### 5. **Observability**
- Comprehensive logging at all levels
- Dual output: stdout (docker-compose logs) + file (./logs/)
- Timestamped log entries with context
- Progress tracking during processing

### 6. **Scalability**
- Stateless services (can run in parallel)
- Database-backed for state persistence
- Worker pool for CPU-efficient parallel processing
- RESTful API for easy integration

### 7. **Code Quality**
- 89.2% test coverage
- 44+ unit and integration tests
- No breaking changes, backward compatible
- Type-safe Go implementation

---

## 🛠 Tech Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| **Language** | Go | 1.25.6 |
| **Web Framework** | Gin | Latest |
| **Database** | MariaDB | 10.11 |
| **ORM** | GORM | Latest |
| **Containerization** | Docker & Docker Compose | 20.10+ |
| **Hot-Reload** | Air | 1.64.5 |
| **Testing** | Go testing + Coverage | Built-in |

---

## 🚀 Local Setup

### Prerequisites

**Option 1: Docker (Recommended)**
```bash
# Check versions
docker --version          # 20.10+
docker-compose --version  # 1.29+
```

**Option 2: Local Development**
```bash
# Check versions
go version        # 1.19+
mariadb --version # 10.11+
```

### Setup Steps

#### Step 1: Clone Repository
```bash
git clone <repository-url>
cd segmentation-api
```

#### Step 2: Start Database

**Using Docker (Recommended):**
```bash
docker-compose up db
```

**Local MariaDB:**
```bash
# macOS
brew services start mariadb

# Or Docker container
docker run -d \
  -e MARIADB_ROOT_PASSWORD=root \
  -e MARIADB_DATABASE=segmentation \
  -e MARIADB_USER=segmentation \
  -e MARIADB_PASSWORD=segmentation \
  -p 3306:3306 \
  mariadb:10.11
```

#### Step 3: Build & Run

**Using Docker Compose:**
```bash
# Terminal 1: Start all services
docker-compose up db api-dev

# Terminal 2: View API logs (verify startup)
docker-compose logs -f api-dev
```

**Local Build:**
```bash
# Build binaries
go build -o api ./cmd/api/
go build -o processor ./cmd/processor/

# Run API server
./api

# Run processor (separate terminal)
./processor
```

#### Step 4: Verify Setup
```bash
# Test API health
curl http://localhost:8080/health

# Expected response
# {"status":"healthy"}
```

### Development with Hot-Reload

**API Development:**
```bash
# Terminal 1: Start database
docker-compose up db

# Terminal 2: Start API with Air hot-reload
docker-compose up db api-dev

# Terminal 3: Edit code
# Files in ./internal/api/* or ./cmd/api/*
# Air automatically recompiles and restarts
```

**Processor Development:**
```bash
# Terminal 1: Start database
docker-compose up db

# Terminal 2: Start processor-dev container
docker-compose --profile dev up db processor-dev

# Terminal 3: Run Air inside container
docker-compose exec processor-dev air -c .air-processor.toml

# Terminal 4: Edit code
# Files in ./internal/processor/* or ./cmd/processor/*
# Air automatically recompiles and reruns
```

### Project Structure

```
segmentation-api/
├── cmd/
│   ├── api/                    # API entry point
│   │   └── main.go
│   └── processor/              # Processor entry point
│       └── main.go
│
├── internal/
│   ├── api/
│   │   ├── handler/            # HTTP handlers
│   │   ├── router.go           # Route definitions
│   │   └── *_test.go
│   │
│   ├── service/                # Business logic
│   │   ├── segmentation.go
│   │   └── *_test.go
│   │
│   ├── repository/             # Data access layer
│   │   ├── segmentation.go     # Interface
│   │   ├── mysql/
│   │   │   └── segmentation.go # Implementation
│   │   └── *_test.go
│   │
│   ├── models/                 # Data structures
│   │   ├── segmentation.go
│   │   └── *_test.go
│   │
│   ├── processor/              # CSV processing
│   │   ├── worker.go
│   │   └── *_test.go
│   │
│   └── logger/                 # Logging
│       └── logger.go
│
├── data/
│   └── data.csv                # Input CSV file
│
├── logs/                       # Output logs (auto-created)
│
├── docker-compose.yaml
├── Dockerfile
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

### Environment Variables

```bash
# API Configuration
API_PORT=8080

# Database Configuration
DATABASE_URL=segmentation:segmentation@tcp(db:3306)/segmentation?charset=utf8mb4&parseTime=true

# Processor Configuration
DATAFILEPATH=/app/data/data.csv
LOG_DIR=/app/logs
```

### Key Docker Commands

```bash
# Start all services
docker-compose up

# Start specific services
docker-compose up db api-dev

# With development profile (includes processor-dev)
docker-compose --profile dev up

# View logs
docker-compose logs -f <service-name>

# Execute command in container
docker-compose exec <service-name> bash

# Stop and clean up
docker-compose down -v
```

### API Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Get user segmentations
curl http://localhost:8080/users/{user_id}/segmentations

# Swagger API Documentation
# Open in browser: http://localhost:8080/swagger/index.html
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run specific package
go test ./internal/service/... -v

# Test coverage
go test ./... -cover
```

### Database Access

**Via Adminer (Web UI):**
- URL: `http://localhost:8081`
- System: MariaDB
- Server: `db` (or `localhost` if local)
- User: `segmentation`
- Password: `segmentation`
- Database: `segmentation`

**Via MySQL CLI:**
```bash
mysql -h localhost -u segmentation -p -D segmentation
# Password: segmentation
```

### Common Workflows

**Process CSV Data:**
```bash
docker-compose up db processor
```

**Development API Changes:**
```bash
docker-compose up db api-dev
# Edit files in internal/api/*
# Air auto-restarts
```

**Development Processor Changes:**
```bash
docker-compose --profile dev up db processor-dev
docker-compose exec processor-dev air -c .air-processor.toml
# Edit files in internal/processor/*
# Air auto-reruns
```

**View Processing Logs:**
```bash
# Docker output
docker-compose logs -f processor-dev

# Log files
tail -f ./logs/2026-02-04T*.log
```

**Run Tests:**
```bash
go test ./...
go test ./internal/service/... -v
go test -run TestName ./...
```

---

## 📚 Additional Documentation

For comprehensive details, see:
- **OLD_README.md** - Extended documentation with all features and troubleshooting

---

## 🤝 Quick Help

**Database won't connect?**
```bash
docker-compose logs db
docker-compose exec db mysql -u root -proot -e "SHOW DATABASES;"
```

**Port already in use?**
```bash
API_PORT=3000 ./api
docker-compose down -v  # Clean up containers/volumes
```

**Tests failing?**
```bash
go test ./... -v
go test -run TestName ./... # Run specific test
```

**Need to rebuild?**
```bash
docker-compose build --no-cache
docker-compose down -v && docker-compose up
```

---

## 👨‍💻 Author

**Charlen Rodrigues**