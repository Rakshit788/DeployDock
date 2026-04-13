
# Vercel Clone - Self-Hosted Deployment Platform

A Go-based, self-hosted deployment platform inspired by Vercel. Deploy web applications from GitHub repositories with automated builds, containerization, and multi-tenant routing.

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go)
![Docker](https://img.shields.io/badge/Docker-20.10-2496ED?style=for-the-badge&logo=docker)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=for-the-badge&logo=postgresql)
![Redis](https://img.shields.io/badge/Redis-7.2-DC382D?style=for-the-badge&logo=redis)
![Nginx](https://img.shields.io/badge/Nginx-Proxy-269539?style=for-the-badge&logo=nginx)

## 📋 Table of Contents

- [Architecture](#architecture)
- [Features](#features)
- [Project Status](#project-status)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [API Endpoints](#api-endpoints)
- [Database Schema](#database-schema)
- [Project Structure](#project-structure)
- [Deployment Workflow](#deployment-workflow)
- [Known Limitations](#known-limitations)
- [Future Roadmap](#future-roadmap)

## 🏗️ Architecture

The platform follows a microservices architecture with async task processing:

```
                       


                    ┌─────────────────────────────────────┐
                    │     DEPLOYMENT PIPELINE             │
                    └─────────────────────────────────────┘

    1. Create Deployment         2. Build Image           3. Deploy Container
    ─────────────────────        ──────────────          ───────────────────
           │                            │                        │
           ▼                            ▼                        ▼
    POST /create-deployment    Worker Service          Deployer Service
           │                     (port 8081)              (async worker)
           │                        │                        │
           ├─► Create DB row        ├─ docker build         ├─ docker run
           │   status: pending      │  from repo URL        │  with VIRTUAL_HOST
           │                        │                       │
           ├─► Enqueue             ├─ Update logs          ├─► nginx detects
           │   build:project        │  in DB               │    VIRTUAL_HOST
           │                        │                       │
           ▼                        ├─ Set url to          ├─► Creates routing
      Redis Queue            │    image name          │    rules
                             │                         │
                             ├─ Enqueue               ├─► Container
                             │  deploy:run            │    starts serving
                             │                         │
                             ▼                        ▼
                        Status: built          Status: deployed
```

### Data Flow

1. **User** → POST `/create-deployment` → **API Service** (port 8080)
2. **API** → Enqueue task → **Redis Queue** 
3. **Worker** → Dequeue `build:project` → Build Docker image → Enqueue deploy task
4. **Deployer** → Dequeue `deploy:run` → Run container with `VIRTUAL_HOST` env
5. **nginx-proxy** → Detects `VIRTUAL_HOST` → Routes requests to container
6. **Container** → Serves application on port 3000 (internal)
7. **User** → Access via direct port OR domain (if DNS configured)

### Services

| Service | Port | Role | Language |
|---------|------|------|----------|
| **API** | 8080 | REST API, OAuth, project/deployment management | Go (Gin) |
| **Worker** | 8081 | Async build processor, Docker image builder | Go |
| **Deployer** | N/A | Async deployment processor, container runner | Go |
| **Nginx** | 80 | Reverse proxy, subdomain routing | (Image-based) |
| **PostgreSQL** | 5432 | Data storage (users, projects, deployments) | N/A |
| **Redis** | 6379 | Task queue broker (asynq) | N/A |

## ✨ Features

### Implemented 
- **GitHub OAuth**: Secure user authentication via GitHub
- **Project Management**: Create projects linked to GitHub repos with custom subdomains
- **Deployment Pipeline**: 
  - Trigger deployments via API
  - Automatic Docker image building from repo
  - Container orchestration and port management
  - Live URL generation
- **Async Task Processing**: Redis queue-based build and deploy workflows
- **Status Tracking**: Real-time deployment status (pending → building → built → deploying → deployed)
- **Database Persistence**: Users, projects, and deployment history
- **Container Networking**: Multi-app routing via Nginx reverse proxy
- **Health Checks**: Service health endpoints and readiness probes

### Limitations 
- Build commands are hardcoded (npm run build) or DB defaults
- GitHub credentials hardcoded in source (should use env vars)
- No automatic database migrations on startup
- Docker builds from Git URLs require git in engine (or use local clone)
- No webhook support for automatic deployments on push
- Windows DNS doesn't resolve nip.io domains without external tools

## 📊 Project Status

**Last Updated**: April 13, 2026

| Component | Status | Notes |
|-----------|--------|-------|
| Core Deployment Flow | ✅ Working | Full pipeline end-to-end |
| API Endpoints | ✅ Working | OAuth, projects, deployments |
| Database Schema | ✅ Ready | Users, projects, deployments tables |
| Docker Integration | ✅ Working | Builds and runs containers |
| Async Queue | ✅ Working | Redis + asynq integration |
| Multi-app Routing | ✅ Working | Nginx-proxy active with VIRTUAL_HOST env vars |
| Container Port Access | ✅ Working | Direct access via 127.0.0.1:random_port |
| Domain Routing | ⚠️ Partial | Nginx routes correctly, Windows DNS lacks nip.io support |
| Worker Service | ✅ Working | Builds and enqueues deploy tasks |
| Deployer Service | ✅ Working | Deploys containers with nginx routing |

## 🛠️ Tech Stack

**Backend**
- Language: Go 1.26
- Web Framework: Gin
- Task Queue: asynq (Redis-backed)
- Async HTTP Client: Standard library

**Infrastructure**
- Containerization: Docker & Docker Compose
- Orchestration: Docker (single host)
- Reverse Proxy: Nginx (nginxproxy/nginx-proxy)
- Database: PostgreSQL 16
- Message Broker: Redis 7.2
- Runtime: Alpine Linux (lightweight images)

**Development**
- Package Manager: Go Modules
- Build Tool: go build (no external build system)
- Testing: (not yet implemented)

## 🚀 Getting Started

### Prerequisites

- Docker Desktop (with Docker Compose)
- Go 1.26+ (for local development)
- Git
- GitHub OAuth application (for authentication)
- 512MB+ free disk space

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/Rakshit788/vercel-clone.git
   cd vercel-clone
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env  # or create manually
   ```
   
   Required variables:
   ```env
   # GitHub OAuth credentials (get from GitHub Settings > Developer Settings > OAuth Apps)
   GITHUB_CLIENT_ID=your_github_client_id
   GITHUB_CLIENT_SECRET=your_github_client_secret
   
   # Optional: custom domain suffix for deployed apps (default: 127.0.0.1)
   BASE_DOMAIN_SUFFIX=127.0.0.1
   ```

3. **Start services**
   ```bash
   docker-compose up --build
   ```

   This starts:
   - PostgreSQL (port 5432)
   - Redis (port 6379)
   - Nginx (port 80)
   - API (port 8080)
   - Worker service
   - Deployer service

4. **Verify services**
   ```bash
   # Check API health
   curl http://localhost:8080/health
   
   # Expected response:
   # {"status":"ok"}
   ```

### Quick Test

1. **Login via GitHub**
   ```bash
   # Browser: http://localhost:8080/auth/github/login
   # Redirects to GitHub → back to your app with user_id in DB
   ```

2. **Create a project**
   ```bash
   curl -X POST http://localhost:8080/create-project \
     -H "Content-Type: application/json" \
     -d '{
       "user_id": 1,
       "name": "my-app",
       "repo_url": "https://github.com/user/repo.git",
       "framework": "next.js"
     }'
   ```

3. **Trigger deployment**
   ```bash
   curl -X POST http://localhost:8080/create-deployment \
     -H "Content-Type: application/json" \
     -d '{"project_id": 1}'
   ```

4. **Check status**
   ```bash
   curl http://localhost:8080/deployments/1/status
   ```

## 📡 API Endpoints

### Authentication
- `GET /auth/github/login` - Redirect to GitHub OAuth
- `GET /auth/github/callback` - GitHub OAuth callback (auto-handled)

### Projects
- `POST /create-project` - Create new project
  ```json
  {
    "user_id": 1,
    "name": "project-name",
    "repo_url": "https://github.com/user/repo.git",
    "framework": "next.js" (optional),
    "subdomain": "custom-name" (optional, auto-generated if not provided)
  }
  ```

### Deployments
- `POST /create-deployment` - Trigger new deployment
  ```json
  {
    "project_id": 1
  }
  ```

- `GET /deployments/:id/status` - Get deployment status
  ```json
  {
    "deployment_id": 1,
    "status": "deployed"
  }
  ```

### Health
- `GET /health` - API health check
- `GET :8081/health` - Worker health check (internal port)

## 🗄️ Database Schema

### users
```sql
id          | SERIAL PRIMARY KEY
github_id   | BIGINT UNIQUE NOT NULL
username    | TEXT NOT NULL
avatar_url  | TEXT
created_at  | TIMESTAMP
```

### projects
```sql
id                  | SERIAL PRIMARY KEY
user_id             | INTEGER REFERENCES users
name                | TEXT NOT NULL
repo_url            | TEXT NOT NULL
framework           | TEXT
subdomain           | TEXT UNIQUE
build_command       | TEXT (default: 'npm run build')
output_directory    | TEXT (default: 'dist')
created_at          | TIMESTAMP
updated_at          | TIMESTAMP
```

### deployments
```sql
id          | SERIAL PRIMARY KEY
project_id  | INTEGER REFERENCES projects
status      | TEXT ('pending','building','built','deploying','deployed','failed')
url         | TEXT (live URL or image name)
commit_sha  | TEXT
logs        | TEXT (build/deploy logs)
started_at  | TIMESTAMP
finished_at | TIMESTAMP
created_at  | TIMESTAMP
```

## 📁 Project Structure

```
vercel-clone/
├── apps/
│   ├── api/                    # REST API service
│   │   ├── cmd/server/         # Entry point
│   │   └── internal/
│   │       ├── auth/           # GitHub OAuth logic
│   │       ├── deployment/     # Deployment API handlers
│   │       └── project/        # Project API handlers
│   ├── worker/                 # Build worker service
│   │   ├── cmd/worker/         # Entry point
│   │   └── internals/
│   │       └── builder.go      # Docker build logic
│   └── deployer/               # Deploy worker service
│       ├── cmd/deployer/       # Entry point
│       └── internals/
│           ├── docker.go       # Container management
│           ├── handler.go      # Deployment handler
│           └── nginx.go        # Nginx config (unused)
├── packages/
│   ├── db/                     # Database connection pool
│   │   ├── pool.go             # pgx pool init
│   │   └── migrations/         # SQL migrations
│   ├── queue/                  # Task queue setup (asynq)
│   ├── redis/                  # Redis client
│   └── config/                 # Config management (empty)
├── infra/
│   ├── docker/
│   │   ├── api.Dockerfile      # API image
│   │   ├── worker.Dockerfile   # Worker image
│   │   └── deployer.Dockerfile # Deployer image
│   └── nginx/                  # Nginx config
├── docker-compose.yaml         # Service orchestration
├── nginx.conf                  # Nginx reverse proxy config
└── README.md                   # This file
```

## 🔄 Deployment Workflow

### Step-by-Step Flow

1. **User triggers deployment**: `POST /create-deployment` with project_id
   - API creates `deployments` row with status `pending`
   - Enqueues `build:project` task to Redis

2. **Worker picks up build task**
   - Sets status → `building`
   - Fetches project repo URL
   - Runs: `docker build -t deploy-{id}:{timestamp} {repo_url}`
   - Captures build logs
   - Sets status → `built`
   - Stores image name in `deployments.url`
   - Enqueues `deploy:run` task

3. **Deployer picks up deploy task**
   - Sets status → `deploying`.{domain} -e VIRTUAL_PORT=3000 {image}`
     - **Important**: Sets VIRTUAL_HOST env var so nginx-proxy recognizes the container
   - Gets random port (3000-9000) and maps it locally
   - Stores container ID and port
   - nginx-proxy automatically detects VIRTUAL_HOST and creates routing rules
   - Sets status → `deployed`
   - Updates `deployments.url` to live URL

4. **User accesses deployment** (two ways)
   - **Direct port**: `http://127.0.0.1:{random_port}` (always works)
   - Direct URL: `http://127.0.0.1:{random_port}`


### Status Transitions

```
pending → building → built → deploying → deployed ✅
   ↓         ↓        ↓         ↓          
 failed   failed    failed    failed    ← Any error state
```

## ⚠️ Known Limitations
Nginx-proxy domain resolution on Windows**
   - nginx-proxy IS active and routes correctly via VIRTUAL_HOST
   - Problem: Windows DNS can't resolve `*.127.0.0.1.nip.io` domains
   - Workaround: Use direct port access `127.0.0.1:port` or add hosts file entry
   - Fix: Use xip.io instead of nip.io, or configure custom domain with hosts entry

2. **Docker builds from Git URLs**
   - Worker runs `docker build -t {image} {git_url}` directly
   - Some Docker engines require git installed for this to work
   - Workaround: Works with Docker BuildKit enabled
   - Fix needed: Clone repo locally first, then build from local context

3. **Hardcoded localhost connections in source code**
   - Some hardcoded values scattered in handler logs and error messages
   - Compose env vars override these at runtime, so not critical
   - Requires local DNS or hosts file entry for `{subdomain}.{domain}`
   - Current workaround: Use direct port URLs (`127.0.0.1:port`)

### Medium Severity
- API returns 200 on JSON bind errors (missing early return)
- Worker panic on queue enqueue failure (should retry gracefully)
- No automatic database migrations on startup
- GitHub OAuth secrets hardcoded in source
- No input validation on endpoints

### Low Priority
- `packages/config/` and `packages/cache/` are empty/unused
- Nginx config generator (`nginx.go`) is dead code
- No rate limiting on API
- No request logging
- No metrics/monitoring

## 🗺️ Future Roadmap

### Phase 1: Core Fixes (High Priority)
- [ ] Move hardcoded secrets to environment variables
- [ ] Fix worker localhost → service name connections
- [ ] Implement proper git clone before Docker build
- [ ] Add database migration runner on startup
- [ ] Better error handling in API handlers

### Phase 2: Features (Medium Priority)
- [ ] GitHub webhooks for automatic deployments on push
- [ ] Deployment rollback functionality
- [ ] Build logs streaming via WebSockets
- [ ] Environment variables per project
- [ ] Custom build commands per project
- [ ] Project listing/retrieval endpoints

### Phase 3: Production Ready (Lower Priority)
- [ ] SSL/TLS certificate management (Let's Encrypt)
- [ ] Multi-user project permissions
- [ ] Deployment history and analytics
- [ ] Docker Compose file generation
- [ ] Buildpack support (auto-detect framework)
- [ ] Pull request preview deployments
- [ ] Horizontal scaling (multiple deployer instances)

### Phase 4: Polish
- [ ] Frontend dashboard UI
- [ ] CLI tool
- [ ] Comprehensive testing
- [ ] API documentation (OpenAPI/Swagger)
- [ ] Monitoring and alerting

## Tech Stack

- **Backend**: Go
- **Framework**: Gin (for the API service)
- **Database**: PostgreSQL with `pgx`
- **Task Queue**: `asynq` (built on Redis)
- **Containerization**: Docker & Docker Compose
- **Reverse Proxy**: Nginx

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go (version 1.21 or higher)
- A GitHub account and an OAuth application

### Installation

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/your-username/vercel-clone.git
    cd vercel-clone
    ```

2.  **Set up environment variables:**

    Create a `.env` file in the root directory and add your GitHub OAuth credentials:
    ```env
    GITHUB_CLIENT_ID=your_github_client_id
    GITHUB_CLIENT_SECRET=your_github_client_secret
    ```

3.  **Run the application:**
    ```bash
    docker-compose up --build
    ```
    This will start all the services. The API will be available at `http://localhost`.

## API Endpoints

- `GET /health`: Health check.
- `GET /auth/github/login`: Redirect to GitHub for OAuth login.
- `GET /auth/github/callback`: Callback URL for GitHub OAuth.
- `POST /create-project`: Create a new project.
- `POST /create-deployment`: Trigger a new deployment for a project.
- `GET /deployments/:id/status`: Get the status of a specific deployment.

## Database Schema

The database consists of three main tables: `users`, `projects`, and `deployments`.

- **users**: Stores user information from GitHub.
- **projects**: Contains details about user projects, including repository URL and build settings.
- **deployments**: Tracks the status and logs of each deployment.

For detailed schema, see `packages/db/migrations/0001_init.sql`.

## Future Scope

- **Dynamic Routing Improvements**: Finalize robust production routing for deployed containers and domains.
- **Real-time Logs**: Stream build logs to the client in real-time using WebSockets.
- **Buildpack Support**: Instead of requiring a `Dockerfile`, use buildpacks to automatically detect the framework and build the application.
- **Preview Deployments**: Create preview deployments for pull requests.

---

*This README was generated with the help of GitHub Copilot.*
