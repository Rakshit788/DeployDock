
# Vercel Clone

This project is a Go-based, self-hosted deployment platform that mimics the core functionalities of Vercel. It allows users to deploy their web projects from GitHub repositories with ease, featuring a microservices architecture for scalability and maintainability.

![Go](https://img.shields.io/badge/Go-1.21-00ADD8?style=for-the-badge&logo=go)
![Docker](https://img.shields.io/badge/Docker-20.10-2496ED?style=for-the-badge&logo=docker)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=for-the-badge&logo=postgresql)
![Redis](https://img.shields.io/badge/Redis-7.2-DC382D?style=for-the-badge&logo=redis)
![Nginx](https://img.shields.io/badge/Nginx-1.25-269539?style=for-the-badge&logo=nginx)

## Architecture

The platform is built on a microservices architecture, consisting of several Go services orchestrated with Docker Compose.

```mermaid
graph TD
    subgraph "User Interaction"
        A[User] -->|Clones Repo/Pushes Code| B(GitHub Repository)
    end

    subgraph "Vercel Clone Platform"
        C(Nginx Reverse Proxy) --> D{API Service}

        subgraph "Async Processing"
            D -->|Enqueues Build Task| E(Redis Queue)
            F(Worker Service) -->|Pulls Task| E
        end

        subgraph "Data Storage"
            D --> G(PostgreSQL Database)
            F --> G
        end

        F -->|Builds & Deploys| H(Docker Container)
    end

    A -->|Creates Project/Deployment| C
    B -->|Webhook (Future)| D
```

- **API Service**: The main entry point for user requests, handling OAuth, project creation, deployment triggers, and status checks.
- **Worker Service**: A background job processor that listens for build tasks from the Redis queue and builds Docker images.
- **Deployer Service**: Consumes deploy tasks and runs built images as Docker containers.
- **Nginx**: Acts as a reverse proxy, directing traffic to the appropriate services.
- **PostgreSQL**: The primary database for storing user, project, and deployment information.
- **Redis**: Used as a message broker for the asynchronous task queue.

## Features

- **GitHub OAuth**: Secure authentication using GitHub accounts.
- **Project Creation**: Link GitHub repositories to create new projects.
- **Automated Builds**: Trigger deployments that automatically clone, build, and containerize your application.
- **Real-time Status**: Monitor the status of your deployments (`pending`, `building`, `deployed`, `failed`).
- **Custom Subdomains**: Assign unique subdomains to your projects.

## What Is Done So Far (April 13, 2026)

### Implemented

- **GitHub OAuth flow is wired end-to-end**: login redirect, callback, GitHub user fetch, and user upsert into `users` table.
- **Project creation is implemented**: `POST /create-project` stores projects and generates sanitized fallback subdomains.
- **Deployment trigger API exists**: `POST /create-deployment` creates a deployment row and enqueues a `build:project` task.
- **Build worker is implemented**: consumes `build:project`, marks deployment `building`, attempts Docker build, writes logs, marks `built`, and enqueues `deploy:run`.
- **Deployer worker is implemented**: consumes `deploy:run`, runs Docker container with `VIRTUAL_HOST`, stores runtime info, and marks deployment `deployed`.
- **Deployment status endpoint exists**: `GET /deployments/:id/status` returns current deployment status.
- **Database schema is in place**: users, projects, deployments tables and indexes are defined in migration SQL.
- **Containerized service setup exists**: API, Postgres, Redis, and deployer are wired in Docker Compose.

### Partially Implemented / Known Gaps

- **Worker runtime config mismatch in containers**: worker currently uses `localhost` for Postgres/Redis, which breaks in Docker Compose networking.
- **Build context bug**: worker calls `docker build` with repo URL directly instead of cloning and building from a local directory.
- **Deployment URL field reuse**: `deployments.url` is used both for image name and final live URL, which mixes two concerns.
- **Hardcoded GitHub OAuth credentials**: client ID/secret are currently embedded in source and should move to environment variables.
- **Migration execution is not automated**: migration SQL exists, but startup does not run migrations.
- **Nginx config generator is unused**: deployer has `GenerateNginxConfig`, but current setup relies on `nginxproxy/nginx-proxy` via container env.
- **Some API error handling needs tightening**: deployment creation path has missing early return on bad JSON and an unnecessary panic on enqueue failure.

### Current Service Status Summary

- **API**: Running and handling auth/project/deployment endpoints.
- **Worker**: Build pipeline logic exists, but runtime/build issues prevent reliable successful builds in current state.
- **Deployer**: Deployment execution logic exists and is connected to queue, pending upstream build/runtime fixes.
- **Infra**: Core Compose stack and database schema are present.

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
