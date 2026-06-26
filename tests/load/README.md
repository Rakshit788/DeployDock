# Load testing DeployDock

The k6 tests run in Docker and write reports to `tests/load/results/`.

- `api-baseline.js` only reads API endpoints. Run this first.
- `deployment-queue.js` creates real deployment jobs, Docker images, and app containers. Use a disposable project/repository.

## Start the platform and run the safe baseline

```powershell
docker compose up --build -d
New-Item -ItemType Directory -Force tests/load/results
docker compose --profile load-test run --rm k6-api-baseline
```

Read `tests/load/results/api-baseline-summary.json` for request count, error rate, median latency, p(95) latency, and virtual users.

## Run a controlled deployment queue test

Use a project whose repository has a valid Dockerfile and is safe to build repeatedly.

```powershell
$env:TEST_PROJECT_ID = '1'
$env:VUS = '4'
$env:ITERATIONS_PER_VU = '1'
docker compose --profile load-test run --rm k6-deployment-queue
```

This creates `VUS × ITERATIONS_PER_VU` jobs. Start at four because builder and deployer concurrency are both currently four. A successful POST only means the job was accepted; inspect completion after the run:

```powershell
docker compose logs worker deployment --tail 200
Invoke-RestMethod http://localhost:8080/api/deployment/list
```

Do not scale the deployment test until this run completes cleanly. Deployment images and containers are not yet automatically cleaned up.
