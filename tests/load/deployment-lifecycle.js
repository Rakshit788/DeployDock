import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Counter, Trend, Rate, Gauge } from 'k6/metrics';

// ─────────────────────────────────────────────────────────
//  Configuration
// ─────────────────────────────────────────────────────────

const BASE_URL = __ENV.BASE_URL || 'http://api:8080';
const POLL_INTERVAL_SEC = Number(__ENV.POLL_INTERVAL || 5);
const MAX_POLL_ATTEMPTS = Number(__ENV.MAX_POLL || 60);  // 5 min max wait

// ─────────────────────────────────────────────────────────
//  Custom Metrics
// ─────────────────────────────────────────────────────────

// Counters
const deploymentsCreated    = new Counter('deployments_created');
const deploymentsSucceeded  = new Counter('deployments_succeeded');
const deploymentsFailed     = new Counter('deployments_failed');
const deploymentsTimedOut   = new Counter('deployments_timed_out');

// Trends (latency distributions)
const createLatency         = new Trend('deployment_create_latency', true);
const e2eLatency            = new Trend('deployment_e2e_latency', true);
const buildLatency          = new Trend('deployment_build_latency', true);

// Rates
const successRate           = new Rate('deployment_success_rate');

// Gauges
const concurrentDeploys     = new Gauge('concurrent_deployments');

// ─────────────────────────────────────────────────────────
//  Load Profile: 4 progressive stages
// ─────────────────────────────────────────────────────────

export const options = {
  scenarios: {
    // Stage 1: Smoke test – 1 VU, 1 deployment
    smoke: {
      executor: 'per-vu-iterations',
      vus: 1,
      iterations: 1,
      maxDuration: '5m',
      startTime: '0s',
      tags: { stage: 'smoke' },
    },

    // Stage 2: Concurrent – 4 VUs (matches builder concurrency)
    concurrent: {
      executor: 'per-vu-iterations',
      vus: Number(__ENV.CONCURRENT_VUS || 4),
      iterations: 1,
      maxDuration: '5m',
      startTime: '5m30s',
      tags: { stage: 'concurrent' },
    },

    // Stage 3: Sustained – ramp up to 8 VUs over time
    sustained: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 4 },
        { duration: '2m',  target: Number(__ENV.SUSTAINED_VUS || 8) },
        { duration: '30s', target: 0 },
      ],
      maxDuration: '8m',
      startTime: '11m',
      tags: { stage: 'sustained' },
    },

    // API health check running throughout
    health_monitor: {
      executor: 'constant-arrival-rate',
      rate: 2,               // 2 req/sec
      timeUnit: '1s',
      duration: '19m',
      preAllocatedVUs: 2,
      maxVUs: 5,
      startTime: '0s',
      tags: { stage: 'monitor' },
      exec: 'healthCheck',
    },
  },

  thresholds: {
    'http_req_failed':                   ['rate<0.05'],             // <5% errors
    'deployment_create_latency':         ['p(95)<2000'],            // create < 2s
    'deployment_e2e_latency':            ['p(95)<300000'],          // full cycle < 5min
    'deployment_success_rate':           ['rate>0.80'],             // >80% success
    'http_req_duration{endpoint:health}':['p(95)<100'],             // health < 100ms
    'checks':                            ['rate>0.90'],
  },
};

// ─────────────────────────────────────────────────────────
//  Setup: Get auth token + find/create test project
// ─────────────────────────────────────────────────────────

export function setup() {
  console.log('╔══════════════════════════════════════════════╗');
  console.log('║  DeployDock Deployment Load Test             ║');
  console.log('╚══════════════════════════════════════════════╝');

  // 1. Get test token
  const tokenRes = http.get(`${BASE_URL}/api/test/token`);
  if (tokenRes.status !== 200) {
    throw new Error(`Cannot get test token: ${tokenRes.status} – ${tokenRes.body}`);
  }
  const token = tokenRes.json('token');
  console.log(`✓ Test token acquired`);

  // 2. Verify health
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, { 'setup: API healthy': (r) => r.status === 200 });

  // 3. Find existing project or tell user to create one
  const headers = { Authorization: `Bearer ${token}` };
  const projectsRes = http.get(`${BASE_URL}/api/project/list`, { headers });
  let projectId = __ENV.TEST_PROJECT_ID;

  if (!projectId) {
    if (projectsRes.status === 200) {
      try {
        const projects = projectsRes.json();
        if (Array.isArray(projects) && projects.length > 0) {
          projectId = String(projects[0].id);
          console.log(`✓ Using existing project ID=${projectId} (${projects[0].name})`);
        }
      } catch (e) {
        // ignore parse error
      }
    }
  }

  if (!projectId) {
    throw new Error('No TEST_PROJECT_ID set and no projects found. Create a project first.');
  }

  console.log(`✓ Target project: ID=${projectId}`);
  console.log(`✓ Poll interval: ${POLL_INTERVAL_SEC}s, Max attempts: ${MAX_POLL_ATTEMPTS}`);

  return { token, projectId: Number(projectId), headers };
}

// ─────────────────────────────────────────────────────────
//  Health Monitor (runs continuously)
// ─────────────────────────────────────────────────────────

export function healthCheck() {
  const res = http.get(`${BASE_URL}/health`, {
    tags: { endpoint: 'health' },
  });
  check(res, {
    'health: status 200': (r) => r.status === 200,
    'health: body ok': (r) => {
      try { return r.json('status') === 'ok'; } catch (_) { return false; }
    },
  });
}

// ─────────────────────────────────────────────────────────
//  Main: Full Deployment Lifecycle
// ─────────────────────────────────────────────────────────

export default function (data) {
  const vuId = __VU;
  const iter = __ITER;

  group('Deployment Lifecycle', () => {
    // ── Step 1: Create deployment ──────────────────────
    let deploymentId;
    const createStart = Date.now();

    group('Create Deployment', () => {
      concurrentDeploys.add(1);

      const payload = JSON.stringify({ project_id: data.projectId });
      const res = http.post(`${BASE_URL}/create-deployment`, payload, {
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${data.token}`,
        },
        tags: { endpoint: 'create_deployment' },
      });

      const latency = Date.now() - createStart;
      createLatency.add(latency);

      const created = check(res, {
        'create: status 200': (r) => r.status === 200,
        'create: has DEPLOYMENT_ID': (r) => {
          try { return Boolean(r.json('DEPLOYMENT_ID')); } catch (_) { return false; }
        },
        'create: status pending': (r) => {
          try { return r.json('status') === 'pending'; } catch (_) { return false; }
        },
      });

      if (created) {
        deploymentId = res.json('DEPLOYMENT_ID');
        deploymentsCreated.add(1);
        console.log(`[VU${vuId}] Created deployment #${deploymentId} (${latency}ms)`);
      } else {
        console.log(`[VU${vuId}] ✗ Failed to create: ${res.status} – ${res.body}`);
        successRate.add(0);
        return;
      }
    });

    // ── Step 2: Poll for completion ────────────────────
    group('Poll Status', () => {
      let finalStatus = 'unknown';
      let attempts = 0;

      while (attempts < MAX_POLL_ATTEMPTS) {
        sleep(POLL_INTERVAL_SEC);
        attempts++;

        const statusRes = http.get(
          `${BASE_URL}/deployments/${deploymentId}/status`,
          {
            headers: { Authorization: `Bearer ${data.token}` },
            tags: { endpoint: 'deployment_status' },
          }
        );

        if (statusRes.status !== 200) {
          console.log(`[VU${vuId}] Poll #${attempts}: HTTP ${statusRes.status}`);
          continue;
        }

        try {
          finalStatus = statusRes.json('status');
        } catch (_) {
          continue;
        }

        // Terminal states
        if (finalStatus === 'deployed' || finalStatus === 'failed') {
          break;
        }

        // Log progress every 6 polls (~30s)
        if (attempts % 6 === 0) {
          console.log(`[VU${vuId}] Deployment #${deploymentId}: status=${finalStatus} (poll ${attempts})`);
        }
      }

      // ── Step 3: Record results ─────────────────────
      const e2e = Date.now() - createStart;
      e2eLatency.add(e2e);

      if (finalStatus === 'deployed') {
        deploymentsSucceeded.add(1);
        successRate.add(1);
        buildLatency.add(e2e);
        console.log(`[VU${vuId}] ✓ Deployment #${deploymentId} DEPLOYED in ${(e2e/1000).toFixed(1)}s`);
      } else if (finalStatus === 'failed') {
        deploymentsFailed.add(1);
        successRate.add(0);
        console.log(`[VU${vuId}] ✗ Deployment #${deploymentId} FAILED after ${(e2e/1000).toFixed(1)}s`);
      } else {
        deploymentsTimedOut.add(1);
        successRate.add(0);
        console.log(`[VU${vuId}] ⏳ Deployment #${deploymentId} TIMED OUT (status=${finalStatus}, ${attempts} polls)`);
      }

      concurrentDeploys.add(0); // reset gauge

      check(null, {
        'deployment completed successfully': () => finalStatus === 'deployed',
        'deployment not timed out': () => attempts < MAX_POLL_ATTEMPTS,
      });
    });
  });
}

// ─────────────────────────────────────────────────────────
//  Teardown: Summary Report
// ─────────────────────────────────────────────────────────

export function teardown(data) {
  console.log('\n╔══════════════════════════════════════════════╗');
  console.log('║  Load Test Complete                          ║');
  console.log('╚══════════════════════════════════════════════╝');

  // List final deployments
  const res = http.get(`${BASE_URL}/api/deployment/list`, {
    headers: { Authorization: `Bearer ${data.token}` },
  });

  if (res.status === 200) {
    try {
      const deployments = res.json();
      console.log(`Total deployments in DB: ${deployments.length}`);

      const statuses = {};
      deployments.forEach((d) => {
        statuses[d.status] = (statuses[d.status] || 0) + 1;
      });

      Object.entries(statuses).forEach(([status, count]) => {
        console.log(`  ${status}: ${count}`);
      });
    } catch (_) {
      // ignore
    }
  }
}
