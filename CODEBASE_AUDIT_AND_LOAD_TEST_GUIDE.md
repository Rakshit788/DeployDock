# DeployDock - Codebase Audit & Load Testing Guide

**Audit Date**: June 22, 2026  
**Status**: 5 Critical Issues Found + Load Testing Ready

---

## 📊 CODEBASE AUDIT RESULTS

### ✅ What's Working

- **Go Compilation**: All Go modules compile without errors
- **Database Migrations**: Properly implemented with advisory locks
- **Docker Configuration**: All Dockerfiles use correct multi-stage builds
- **Worker Service**: Successfully builds Docker images with timeout handling
- **API Structure**: All endpoints exist and respond

### 🚨 Critical Issues Found

| Priority | Issue | Location | Impact |
|----------|-------|----------|--------|
| **P0** | **Deployer service missing from docker-compose** | [docker-compose.yaml](docker-compose.yaml) | Deployments stuck in "built" status forever |
| **P0** | **No JWT token generation** | [api/internal/auth/github.go](apps/api/internal/auth/github.go) | Authentication completely broken |
| **P1** | **No user validation on project creation** | [api/internal/project/handler.go](apps/api/internal/project/handler.go) | Users can create projects as other users |
| **P1** | **No user filtering on deployments** | [api/internal/deployment/handler.go](apps/api/internal/deployment/handler.go) | Users see all deployments (data leak) |
| **P1** | **Hardcoded Redis in deployer** | [deployer/cmd/deployer/main.go](apps/deployer/cmd/deployer/main.go) | Ignores REDIS_ADDR env variable |

### 🔒 Security Issues

- **Multi-tenant isolation broken**: Users accessing other users' data
- **No API authentication tokens**: Frontend cannot authenticate requests
- **GitHub OAuth scope mismatch**: Frontend requests "repo" scope, backend expects "user"

---

## 🚀 BEFORE LOAD TESTING

**You MUST fix P0 issues first, otherwise load tests won't work:**

1. Add deployer to docker-compose.yaml
2. Implement JWT token generation in GitHub OAuth flow
3. Add user validation/filtering

---

## 📈 LOAD TESTING SETUP

Your project has **k6** load testing configured with two test suites:

### Test 1: API Baseline (Safe - Read-Only)
Tests API health, project list, and deployment list endpoints.

```powershell
# Setup
docker compose up --build -d
New-Item -ItemType Directory -Force tests/load/results

# Run baseline test
docker compose --profile load-test run --rm k6-api-baseline
```

**What it tests:**
- Health endpoint
- Project list (read-only)
- Deployment list (read-only)
- Load: 5 virtual users for 30 seconds

**Expected output:** `tests/load/results/api-baseline-summary.json`

---

### Test 2: Deployment Queue (Advanced - Creates Real Jobs)
Creates actual deployment jobs. Use a test project with a valid Dockerfile.

```powershell
# Set environment variables
$env:TEST_PROJECT_ID = '1'              # Must exist in your database
$env:VUS = '4'                          # Virtual users (max 4 - builder concurrency)
$env:ITERATIONS_PER_VU = '1'            # Jobs per user (start low!)
$env:BASE_URL = 'http://api:8080'       # Optional

# Run deployment queue test
docker compose --profile load-test run --rm k6-deployment-queue
```

**Important Notes:**
- Start with VUS=4, ITERATIONS_PER_VU=1 (= 4 total jobs)
- Builder and deployer concurrency are both 4
- Each job creates a Docker image and container
- **Containers/images are NOT auto-cleaned up** - clean manually after
- Don't increase load until deployment completes

**After test, check completion:**
```powershell
docker compose logs worker deployer --tail 200
Invoke-RestMethod http://localhost:8080/api/deployment/list
```

---

## 📊 LOAD TEST STAGES

### Stage 1: Baseline Validation (Day 1)
```powershell
# Run read-only tests to verify API stability
docker compose up --build -d
docker compose --profile load-test run --rm k6-api-baseline

# Check results
Get-Content tests/load/results/api-baseline-summary.json | ConvertFrom-Json | Format-Table
```

**Success Criteria:**
- Error rate < 1%
- p(95) latency < 500ms
- All health checks pass

---

### Stage 2: Single Deployment Job (Day 2)
```powershell
# Ensure test project exists (ID=1)
Invoke-RestMethod http://localhost:8080/api/project/list?user_id=1

# Create 1 deployment
$env:TEST_PROJECT_ID = '1'
$env:VUS = '1'
$env:ITERATIONS_PER_VU = '1'
docker compose --profile load-test run --rm k6-deployment-queue

# Monitor
docker compose logs worker deployer --tail 100
```

**Success Criteria:**
- Deployment accepted (HTTP 200)
- Worker builds image successfully
- Deployer runs container

---

### Stage 3: 4 Concurrent Deployments (Day 3+)
```powershell
# Match builder concurrency limit
$env:TEST_PROJECT_ID = '1'
$env:VUS = '4'
$env:ITERATIONS_PER_VU = '1'
docker compose --profile load-test run --rm k6-deployment-queue

# Monitor
docker compose logs worker --tail 200
docker compose logs deployer --tail 200

# Verify all 4 deployments completed
Invoke-RestMethod http://localhost:8080/api/deployment/list | ConvertTo-Json
```

---

### Stage 4: Stress Testing (When Ready)
```powershell
# Increase load beyond concurrency limit
$env:VUS = '10'
$env:ITERATIONS_PER_VU = '2'
docker compose --profile load-test run --rm k6-deployment-queue

# Monitor queue depth and latency
docker compose logs redis
```

---

## 🧹 Cleanup After Load Tests

```powershell
# List running containers
docker ps --filter "label=deployer_container" -q

# List built images
docker images --filter "reference=*deployer*" -q

# Remove images manually
docker rmi <image-id>

# Remove containers manually
docker rm <container-id>
```

---

## 📋 Test Configuration Details

### api-baseline.js
- **Duration**: 60 seconds
- **Stages**: 0→5 VUs (15s), hold 5 (30s), 5→0 (15s)
- **Thresholds**:
  - Error rate < 1%
  - p(95) latency < 500ms
  - Check pass rate > 99%

### deployment-queue.js
- **Duration**: 2 minutes max
- **Concurrency**: configurable via $env:VUS
- **Thresholds**:
  - Error rate < 1%
  - p(95) latency < 1000ms
  - Check pass rate > 99%

---

## 🐛 Troubleshooting

### "Connection refused" on docker compose run
→ Ensure `docker compose up -d` completed healthchecks first

### Deployments stuck in "built" status
→ **Deployer service missing** - Add to docker-compose.yaml

### "invalid_grant" errors from GitHub
→ JWT token not being generated - Check auth flow implementation

### Container image build fails
→ Worker tries to build from GitHub URL - Check if repo is accessible

### "too many open files" errors
→ Increase ulimit: `ulimit -n 65536`

---

## 📈 Monitoring During Load Tests

### Watch logs in real-time
```powershell
# API logs
docker compose logs -f api

# Worker logs
docker compose logs -f worker

# All services
docker compose logs -f
```

### Check metrics
```powershell
# Redis queue depth
docker exec vercel_redis redis-cli DBSIZE

# PostgreSQL connections
docker exec vercel_postgres psql -U postgres -d vercel_clone -c "SELECT * FROM pg_stat_activity;"
```

---

## ✅ Pre-Load-Test Checklist

- [ ] All P0 issues fixed
- [ ] Deployer added to docker-compose.yaml
- [ ] JWT token generation implemented
- [ ] User validation added to endpoints
- [ ] Test project created (ID=1)
- [ ] Test repository has valid Dockerfile
- [ ] Docker socket accessible to nginx-proxy
- [ ] Sufficient disk space for images/containers
- [ ] Load test results directory exists

---

## 📝 Next Steps

1. **Fix critical issues** in this order:
   - Add deployer service to docker-compose
   - Implement JWT token generation
   - Add user validation filters

2. **Run baseline test** (Stage 1)
3. **Test single deployment** (Stage 2)
4. **Increase load gradually** (Stage 3+)
5. **Monitor and tune** performance

---

*For detailed code analysis, see the individual files mentioned in the issues table.*
