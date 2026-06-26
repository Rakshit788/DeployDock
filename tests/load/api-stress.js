import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend, Rate } from 'k6/metrics';

// ─────────────────────────────────────────────────────────
//  API Stress Test – Push the API to its breaking point
// ─────────────────────────────────────────────────────────

const BASE_URL = __ENV.BASE_URL || 'http://api:8080';

// Custom Metrics
const healthLatency  = new Trend('health_latency', true);
const projectLatency = new Trend('project_list_latency', true);
const deployLatency  = new Trend('deployment_list_latency', true);
const errorRate      = new Rate('error_rate');
const errorCount     = new Counter('error_count');

export const options = {
  scenarios: {
    // Stage 1: Warm up
    warmup: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 10 },
        { duration: '20s', target: 10 },
      ],
      startTime: '0s',
      tags: { phase: 'warmup' },
    },

    // Stage 2: Ramp to 50 concurrent users
    ramp: {
      executor: 'ramping-vus',
      startVUs: 10,
      stages: [
        { duration: '30s', target: 50 },
        { duration: '30s', target: 50 },
      ],
      startTime: '30s',
      tags: { phase: 'ramp' },
    },

    // Stage 3: Spike to 100 concurrent users
    spike: {
      executor: 'ramping-vus',
      startVUs: 50,
      stages: [
        { duration: '15s', target: 100 },
        { duration: '30s', target: 100 },
        { duration: '15s', target: 0 },
      ],
      startTime: '90s',
      tags: { phase: 'spike' },
    },

    // Stage 4: Soak at moderate load
    soak: {
      executor: 'constant-arrival-rate',
      rate: 30,          // 30 req/s
      timeUnit: '1s',
      duration: '60s',
      preAllocatedVUs: 20,
      maxVUs: 40,
      startTime: '150s',
      tags: { phase: 'soak' },
      exec: 'soakTest',
    },
  },

  thresholds: {
    'http_req_failed':                ['rate<0.05'],      // <5% errors
    'http_req_duration':              ['p(95)<1000'],     // p95 < 1s
    'health_latency':                 ['p(95)<200'],      // health < 200ms
    'project_list_latency':           ['p(95)<500'],      // project list < 500ms
    'deployment_list_latency':        ['p(95)<500'],      // deployment list < 500ms
    'error_rate':                     ['rate<0.05'],      // <5% errors
    'checks':                         ['rate>0.95'],
  },
};

// ─────────────────────────────────────────────────────────
//  Setup: Get auth token
// ─────────────────────────────────────────────────────────

export function setup() {
  console.log('╔══════════════════════════════════════════════╗');
  console.log('║  DeployDock API Stress Test                  ║');
  console.log('╚══════════════════════════════════════════════╝');

  const tokenRes = http.get(`${BASE_URL}/api/test/token`);
  if (tokenRes.status !== 200) {
    throw new Error(`Cannot get test token: ${tokenRes.status}`);
  }
  const token = tokenRes.json('token');
  console.log('✓ Test token acquired');

  return { token, headers: { Authorization: `Bearer ${token}` } };
}

// ─────────────────────────────────────────────────────────
//  Main Default Function (warmup, ramp, spike)
// ─────────────────────────────────────────────────────────

export default function (data) {
  // Health check
  const h = http.get(`${BASE_URL}/health`, { tags: { endpoint: 'health' } });
  healthLatency.add(h.timings.duration);
  const hOk = check(h, {
    'health: 200': (r) => r.status === 200,
  });
  if (!hOk) { errorRate.add(1); errorCount.add(1); } else { errorRate.add(0); }

  // Project list (authenticated)
  const p = http.get(`${BASE_URL}/api/project/list`, {
    headers: data.headers,
    tags: { endpoint: 'project_list' },
  });
  projectLatency.add(p.timings.duration);
  const pOk = check(p, {
    'project list: 200': (r) => r.status === 200,
  });
  if (!pOk) { errorRate.add(1); errorCount.add(1); } else { errorRate.add(0); }

  // Deployment list (authenticated)
  const d = http.get(`${BASE_URL}/api/deployment/list`, {
    headers: data.headers,
    tags: { endpoint: 'deployment_list' },
  });
  deployLatency.add(d.timings.duration);
  const dOk = check(d, {
    'deployment list: 200': (r) => r.status === 200,
  });
  if (!dOk) { errorRate.add(1); errorCount.add(1); } else { errorRate.add(0); }

  sleep(0.5);
}

// ─────────────────────────────────────────────────────────
//  Soak Function (constant arrival rate)
// ─────────────────────────────────────────────────────────

export function soakTest(data) {
  // Randomly pick endpoint
  const endpoints = [
    { url: `${BASE_URL}/health`, tag: 'health', auth: false },
    { url: `${BASE_URL}/api/project/list`, tag: 'project_list', auth: true },
    { url: `${BASE_URL}/api/deployment/list`, tag: 'deployment_list', auth: true },
  ];

  const ep = endpoints[Math.floor(Math.random() * endpoints.length)];
  const opts = { tags: { endpoint: ep.tag } };
  if (ep.auth) opts.headers = data.headers;

  const res = http.get(ep.url, opts);

  const ok = check(res, {
    [`${ep.tag}: 200`]: (r) => r.status === 200,
  });

  if (!ok) { errorRate.add(1); errorCount.add(1); } else { errorRate.add(0); }
}

// ─────────────────────────────────────────────────────────
//  Teardown
// ─────────────────────────────────────────────────────────

export function teardown() {
  console.log('\n╔══════════════════════════════════════════════╗');
  console.log('║  API Stress Test Complete                    ║');
  console.log('╚══════════════════════════════════════════════╝');
}
